package changes

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sirupsen/logrus"
)

type ChangesOptions struct {
	utils.BaseOptions
	// AuthorsRegex string
	Since string
	Until string
}

type LinesChanges struct {
	New      int
	Refactor int
	Churn    int
	Helper   int
}

type AuthorLines struct {
	Author string
	Lines  LinesChanges
}

type ChangesFileResult struct {
	CommitId string
	FilePath string
	ChangesResult
}

type ChangesResult struct {
	TotalLines     LinesChanges
	TotalFiles     int
	TotalCommits   int
	authorLinesMap map[string]LinesChanges // temporary map used during processing
	AuthorsLines   []AuthorLines
	analysisTime   time.Duration
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

	logrus.Debugf("Analysing changes in branch %s from %s to %s", opts.Branch, opts.Since, opts.Until)

	fre, err := regexp.Compile(opts.FilesRegex)
	if err != nil {
		return result, errors.New("file filter regex is invalid. err=" + err.Error())
	}

	// MAP REDUCE - analyse files in parallel goroutines
	// we need to start workers in the reverse order so that all the chain
	// is prepared when submitting tasks to avoid deadlocks
	// nrWorkers := runtime.NumCPU() - 1
	nrWorkers := 1
	logrus.Debugf("Preparing a pool of workers to process file analysis in parallel")
	analyseFileInputChan := make(chan analyseFileRequest, 5000)
	analyseFileOutputChan := make(chan ChangesFileResult, 5000)
	analyseFileErrChan := make(chan error, nrWorkers)

	// REDUCE - summarise counters (STEP 3/3)
	var summaryWorkerWaitGroup sync.WaitGroup
	summaryWorkerWaitGroup.Add(1)
	go func() {
		defer summaryWorkerWaitGroup.Done()
		logrus.Debugf("Counting total lines changed per author")
		for fileResult := range analyseFileOutputChan {
			result.TotalFiles += fileResult.TotalFiles
			result.TotalLines.Churn += fileResult.TotalLines.Churn
			result.TotalLines.Helper += fileResult.TotalLines.Helper
			result.TotalLines.New += fileResult.TotalLines.New
			result.TotalLines.Refactor += fileResult.TotalLines.Refactor
			for author := range fileResult.authorLinesMap {
				fileAuthorLines := fileResult.authorLinesMap[author]
				resultAuthorLines := result.authorLinesMap[author]
				resultAuthorLines.Churn += fileAuthorLines.Churn
				resultAuthorLines.Helper += fileAuthorLines.Helper
				resultAuthorLines.New += fileAuthorLines.New
				resultAuthorLines.Refactor += fileAuthorLines.Refactor
			}
			progressInfo.CompletedTasks += 1
			progressInfo.CompletedTotalTime += result.analysisTime
			progressInfo.Message = fmt.Sprintf("%s", fileResult.FilePath)
			if len(progressChan) < 1 {
				progressChan <- progressInfo
			}
		}

		logrus.Debugf("Sorting and preparing summary for each author")
		authorsLines := make([]AuthorLines, 0)
		for author := range result.authorLinesMap {
			lines := result.authorLinesMap[author]
			authorsLines = append(authorsLines, AuthorLines{Author: author, Lines: lines})
		}

		sort.Slice(authorsLines, func(i, j int) bool {
			ai := authorsLines[i].Lines
			aj := authorsLines[j].Lines
			return ai.Churn+ai.Helper+ai.New+ai.Refactor > aj.Churn+aj.Helper+aj.New+aj.Refactor
		})
		result.AuthorsLines = authorsLines
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
		result.TotalCommits = len(commits)

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

	// fmt.Printf("SUMMARY: %v\n", result)

	return result, nil
}

// this will be run by multiple goroutines
func analyseFileChangesWorker(analyseFileInputChan <-chan analyseFileRequest, analyseFileOutputChan chan<- ChangesFileResult, analyseFileErrChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	diffMatcher := diffmatchpatch.New()
	// diffMatcher.MatchDistance = 200

	for req := range analyseFileInputChan {

		fmt.Printf(">>>>%s %s\n", req.commitId, req.filePath)

		startTime := time.Now()
		changesFileResult := ChangesFileResult{
			CommitId: req.commitId,
			FilePath: req.filePath,
			ChangesResult: ChangesResult{
				TotalLines:     LinesChanges{},
				authorLinesMap: make(map[string]LinesChanges, 0),
				AuthorsLines:   []AuthorLines{},
			},
		}

		finfo, err := os.Stat(fmt.Sprintf("%s/%s", req.repoDir, req.filePath))
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Couldn't open file. file=%s. err=%s", req.filePath, err))
			break
		}
		if finfo.Size() > 80000 {
			logrus.Debugf("Ignoring file because it's too big. file=%s, size=%d", req.filePath, finfo.Size())
			continue
		}

		// blame previous version of the file (so we can compare from->to contents)
		prevCommitId, err := utils.ExecPreviousCommitId(req.repoDir, req.commitId)
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Error on getting prev commit id. err=%s", err))
			break
		}

		if prevCommitId == "" {
			logrus.Debugf("No previous commit found. Skipping. commitId=%s", req.commitId)
			continue
		}

		filePrevBlame, err := utils.ExecGitBlame(req.repoDir, req.filePath, prevCommitId)
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Error on git blame prev. file=%s. err=%s", req.filePath, err))
			break
		}
		var prevBuffer bytes.Buffer
		for _, line := range filePrevBlame {
			prevBuffer.WriteString(line.LineContents)
			prevBuffer.WriteString("\n")
		}
		filePrevContents := prevBuffer.String()

		// blame current version of the file
		fileCurBlame, err := utils.ExecGitBlame(req.repoDir, req.filePath, req.commitId)
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Error on git blame cur. file=%s. err=%s", req.filePath, err))
			break
		}
		var curBuffer bytes.Buffer
		for _, line := range fileCurBlame {
			curBuffer.WriteString(line.LineContents)
			curBuffer.WriteString("\n")
		}
		fileCurContents := curBuffer.String()

		// diff both versions of the file
		// diffs := diffMatcher.DiffMain(filePrevContents, fileCurContents, false)

		fileAdmp, fileBdmp, dmpStrings := diffMatcher.DiffLinesToChars(filePrevContents, fileCurContents)
		diffs := diffMatcher.DiffMain(fileAdmp, fileBdmp, false)
		diffs = diffMatcher.DiffCharsToLines(diffs, dmpStrings)
		diffs = diffMatcher.DiffCleanupSemantic(diffs)

		fmt.Printf("curCommitId=%s; prevCommitId=%s\n", req.commitId, prevCommitId)

		fmt.Println("###############")
		// fmt.Println(diffMatcher.DiffPrettyText(diffs))
		fmt.Println("###############")

		changesFileResult.TotalFiles += 1
		// for each line, classify change type
		for li, diff := range diffs {
			// fmt.Printf("%s - %s | %s -> %s\n", diff.Text, diff.Type.String(), filePrevBlame[li], fileCurBlame[li])
			fmt.Printf("%s %s\n", diff.Type.String(), diff.Text)
			if strings.Trim(diff.Text, " ") == "" {
				continue
			}
			changesFileResult.TotalLines.Churn += 1
			// changesFileResult.TotalLines.Helper += 1
			// changesFileResult.TotalLines.New += 1
			// changesFileResult.TotalLines.Refactor += 1
			authorLine, ok := changesFileResult.authorLinesMap[fileCurBlame[li].AuthorName]
			if !ok {
				authorLine = LinesChanges{}
			}
			authorLine.Churn += 1
			// authorLine.Helper += 1
			// authorLine.New += 1
			// authorLine.Refactor += 1
			changesFileResult.authorLinesMap[fileCurBlame[li].AuthorName] = authorLine
		}
		changesFileResult.analysisTime = time.Since(startTime)
		analyseFileOutputChan <- changesFileResult
		// time.Sleep(1 * time.Second)
		// fmt.Printf("Time spent: %s\n", time.Since(startTime))
	}

}
