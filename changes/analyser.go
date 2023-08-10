package changes

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
)

type ChangesOptions struct {
	utils.BaseOptions
	// AuthorsRegex string
	Since string
	Until string
}

type LinesChanges struct {
	/* New lines found in commits */
	New int

	/* Lines changed in commits. If the same line is changed in two commits, for example, it will count as two changes. This is the sum of RefactorOwn, RefactorOther, ChurnOwn and ChurnOther */
	Changes int

	/* Lines changed after a while in which the author of the previous version was the same person */
	RefactorOwn int
	/* Lines changed after a while in which the author of the previous version was another person */
	RefactorOther int
	/* Lines you owned that were changed by another person after a while. When adding RefactorOther to someone, the author of the previous version of the line will have this counter incremented */
	RefactorReceived int

	/* Lines changed in a short term in which the author of the previous version was the same person */
	ChurnOwn int
	/* Lines changed in a short term in which the author of the previous version was another person */
	ChurnOther int
	/* Lines you owned that were changed by another person in a short term. When adding ChurnOther to someone, the author of the previous version of the line will have this counter incremented */
	ChurnReceived int

	/* Sum of age of lines in the moment they are changed. AgeSum/Changes gives you the average survival duration of a line before it's changed by someone */
	AgeSum time.Duration
}

type AuthorLines struct {
	AuthorName string
	Lines      LinesChanges
}

type ChangesFileResult struct {
	CommitId string
	FilePath string
	ChangesResult
}

type ChangesResult struct {
	/* Lines change stats */
	TotalLines LinesChanges
	/* Total files changed in the different commits. If the same file is changed in two commits, for example, it will count as two. */
	TotalFiles int
	/* Number of commits analysed */
	TotalCommits   int
	authorLinesMap map[string]LinesChanges // temporary map used during processing
	/* Change stats per author */
	AuthorsLines []AuthorLines
	analysisTime time.Duration
}

type analyseFileRequest struct {
	repoDir  string
	commitId string
	filePath string
}

func AnalyseChanges(opts ChangesOptions, progressChan chan<- utils.ProgressInfo) (ChangesResult, error) {
	result := ChangesResult{
		TotalLines:     LinesChanges{},
		authorLinesMap: make(map[string]LinesChanges, 0),
		AuthorsLines:   make([]AuthorLines, 0)}

	progressInfo := utils.ProgressInfo{}

	if opts.Branch == "" {
		return ChangesResult{}, fmt.Errorf("opts.Branch is required")
	}

	logrus.Debugf("Analysing changes in branch %s from %s to %s", opts.Branch, opts.Since, opts.Until)

	fre, err := regexp.Compile(opts.FilesRegex)
	if err != nil {
		return result, errors.New("file filter regex is invalid. err=" + err.Error())
	}

	// MAP REDUCE - analyse files in parallel goroutines
	// we need to start workers in the reverse order so that all the chain
	// is prepared when submitting tasks to avoid deadlocks
	nrWorkers := runtime.NumCPU() - 1
	// nrWorkers := 1
	logrus.Debugf("Preparing a pool of workers to process file analysis in parallel")
	analyseFileInputChan := make(chan analyseFileRequest, 5000)
	analyseFileOutputChan := make(chan ChangesFileResult, 5000)
	analyseFileErrChan := make(chan error, nrWorkers)

	// REDUCE - summarise counters (STEP 3/3)
	var summaryWorkerWaitGroup sync.WaitGroup
	summaryWorkerWaitGroup.Add(1)
	go func() {
		defer summaryWorkerWaitGroup.Done()

		commitsWithFiles := make(map[string]bool, 0)

		logrus.Debugf("Counting total lines changed per author")
		for fileResult := range analyseFileOutputChan {
			commitsWithFiles[fileResult.CommitId] = true
			result.TotalFiles += fileResult.TotalFiles
			result.TotalLines = sumLinesChanges(result.TotalLines, fileResult.TotalLines)
			for author := range fileResult.authorLinesMap {
				fileAuthorLines := fileResult.authorLinesMap[author]
				authorLines := result.authorLinesMap[author]
				authorLines = sumLinesChanges(authorLines, fileAuthorLines)
				result.authorLinesMap[author] = authorLines
			}
			progressInfo.CompletedTasks += 1
			progressInfo.CompletedTotalTime += result.analysisTime
			progressInfo.Message = fmt.Sprintf("%s", fileResult.FilePath)
			if progressChan != nil {
				progressChan <- progressInfo
			}
		}

		result.TotalCommits = len(commitsWithFiles)
	}()

	// MAP - start analyser workers (STEP 2/3)
	var analysisWorkersWaitGroup sync.WaitGroup
	for i := 0; i < nrWorkers; i++ {
		analysisWorkersWaitGroup.Add(1)
		go analyseFileChangesWorker(analyseFileInputChan, analyseFileOutputChan, analyseFileErrChan, &analysisWorkersWaitGroup)
	}
	logrus.Debugf("Launched %d workers for analysis", nrWorkers)

	// MAP - submit tasks (STEP 1/3)
	// for each commit between 'from' and 'to' date, discover changed files and submit then for analysis
	go func() {

		totalFiles := 0
		progressInfo.TotalTasksKnown = false

		commits, err := utils.ExecCommitsInRange(opts.RepoDir, opts.Branch, opts.Since, opts.Until)
		if err != nil {
			logrus.Errorf("Error getting commits. err=%s", err)
			panic(5)
		}

		for _, commitId := range commits {
			// fmt.Printf(">>>>%s\n", commitId)
			files, err := utils.ExecDiffTree(opts.RepoDir, commitId)
			if err != nil {
				logrus.Errorf("Error getting files changed in commit. err=%s", err)
				panic(5)
			}

			for _, fileName := range files {
				if strings.Trim(fileName, " ") == "" || !fre.MatchString(fileName) {
					// logrus.Debugf("Ignoring file %s", file.Name)
					continue
				}
				totalFiles += 1
				progressInfo.TotalTasks += 1
				analyseFileInputChan <- analyseFileRequest{repoDir: opts.RepoDir, filePath: fileName, commitId: commitId}
			}
		}

		// finished publishing request messages
		logrus.Debugf("%d files scheduled for analysis", totalFiles)
		logrus.Debug("Task submission worker finished")
		close(analyseFileInputChan)

		progressInfo.TotalTasksKnown = true
		if progressChan != nil {
			progressChan <- progressInfo
		}
	}()

	analysisWorkersWaitGroup.Wait()
	logrus.Debug("Analysis workers finished")
	close(analyseFileOutputChan)
	close(analyseFileErrChan)

	for workerErr := range analyseFileErrChan {
		logrus.Errorf("Error during analysis. err=%s", workerErr)
		panic(2)
	}

	summaryWorkerWaitGroup.Wait()
	logrus.Debug("Summary worker finished")

	logrus.Debugf("Preparing summary for each author")
	authorsLines := make([]AuthorLines, 0)
	for author := range result.authorLinesMap {
		lines := result.authorLinesMap[author]
		authorsLines = append(authorsLines, AuthorLines{AuthorName: author, Lines: lines})
	}

	sort.Slice(authorsLines, func(i, j int) bool {
		ai := authorsLines[i].Lines
		aj := authorsLines[j].Lines
		return ai.New+ai.Changes > aj.New+aj.Changes
	})
	result.AuthorsLines = authorsLines

	// fmt.Printf("SUMMARY: %v\n", result)

	return result, nil
}

func sumLinesChanges(changes1 LinesChanges, changes2 LinesChanges) LinesChanges {
	changes1.Changes += changes2.Changes
	changes1.ChurnOther += changes2.ChurnOther
	changes1.ChurnOwn += changes2.ChurnOwn
	changes1.ChurnReceived += changes2.ChurnReceived
	changes1.New += changes2.New
	changes1.RefactorOther += changes2.RefactorOther
	changes1.RefactorOwn += changes2.RefactorOwn
	changes1.RefactorReceived += changes2.RefactorReceived
	changes1.AgeSum += changes2.AgeSum
	return changes1
}
