package ownership

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
)

type OwnershipOptions struct {
	utils.BaseOptions
	When string
}

type AuthorLines struct {
	AuthorName string
	OwnedLines int
}
type OwnershipResult struct {
	TotalFiles     int
	TotalLines     int
	authorLinesMap map[string]int // temporary map used during processing
	AuthorsLines   []AuthorLines
	CommitId       string
	FilePath       string
	blameTime      time.Duration
}

type analyseFileRequest struct {
	repoDir  string
	filePath string
	commitId string
}

func AnalyseCodeOwnership(opts OwnershipOptions, progressChan chan<- utils.ProgressInfo) (OwnershipResult, error) {
	result := OwnershipResult{TotalLines: 0, authorLinesMap: make(map[string]int, 0), AuthorsLines: make([]AuthorLines, 0)}

	progressInfo := utils.ProgressInfo{}

	fre, err := regexp.Compile(opts.FilesRegex)
	if err != nil {
		return result, errors.New("file filter regex is invalid. err=" + err.Error())
	}

	logrus.Debugf("Analysing branch %s at %s", opts.Branch, opts.When)

	commitId, err := utils.ExecGetCommitAtDate(opts.RepoDir, opts.Branch, opts.When)
	if err != nil {
		return result, err
	}
	result.CommitId = commitId

	// MAP REDUCE - analyse files in parallel goroutines
	// we need to start workers in the reverse order so that all the chain
	// is prepared when submitting tasks to avoid deadlocks
	nrWorkers := runtime.NumCPU() - 1
	// nrWorkers := 1
	logrus.Debugf("Preparing a pool of workers to process file analysis in parallel")
	analyseFileInputChan := make(chan analyseFileRequest, 5000)
	analyseFileOutputChan := make(chan OwnershipResult, 5000)
	analyseFileErrChan := make(chan error, nrWorkers)

	// REDUCE - summarise counters (STEP 3/3)
	var summaryWorkerWaitGroup sync.WaitGroup
	summaryWorkerWaitGroup.Add(1)
	go func() {
		defer summaryWorkerWaitGroup.Done()
		logrus.Debugf("Counting total lines owned per author")
		for fileResult := range analyseFileOutputChan {
			result.TotalFiles += fileResult.TotalFiles
			result.TotalLines += fileResult.TotalLines
			for author := range fileResult.authorLinesMap {
				authorLines := fileResult.authorLinesMap[author]
				result.authorLinesMap[author] = authorLines + result.authorLinesMap[author]
			}
			progressInfo.CompletedTasks += 1
			progressInfo.CompletedTotalTime += fileResult.blameTime
			progressInfo.Message = fmt.Sprintf("%s (%dms)", fileResult.FilePath, fileResult.blameTime.Milliseconds())
			if progressChan != nil {
				progressChan <- progressInfo
			}
		}

		logrus.Debugf("Sorting and preparing summary for each author")
		authorsLines := make([]AuthorLines, 0)
		for author := range result.authorLinesMap {
			lines := result.authorLinesMap[author]
			authorsLines = append(authorsLines, AuthorLines{AuthorName: author, OwnedLines: lines})
		}

		sort.Slice(authorsLines, func(i, j int) bool {
			return authorsLines[i].OwnedLines > authorsLines[j].OwnedLines
		})
		result.AuthorsLines = authorsLines
	}()

	// MAP - start analyser workers (STEP 2/3)
	var analysisWorkersWaitGroup sync.WaitGroup
	for i := 0; i < nrWorkers; i++ {
		analysisWorkersWaitGroup.Add(1)
		go blameFileWorker(analyseFileInputChan, analyseFileOutputChan, analyseFileErrChan, &analysisWorkersWaitGroup)
	}
	logrus.Debugf("Launched %d workers for analysis", nrWorkers)

	// MAP - submit tasks (STEP 1/3)
	go func() {
		logrus.Debugf("Scheduling files for analysis. filesRegex=%s", opts.FilesRegex)
		totalFiles := 0
		progressInfo.TotalTasksKnown = false
		files, err := utils.ExecListTree(opts.RepoDir, commitId)
		// tree, err := commitObj.Tree()
		if err != nil {
			logrus.Errorf("Error getting commit tree. err=%s", err)
			panic(5)
		}

		for _, fileName := range files {
			if !fre.MatchString(fileName) {
				// logrus.Debugf("Ignoring file %s", file.Name)
				continue
			}
			totalFiles += 1
			progressInfo.TotalTasks += 1
			analyseFileInputChan <- analyseFileRequest{repoDir: opts.RepoDir, filePath: fileName, commitId: commitId}
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
func blameFileWorker(analyseFileInputChan <-chan analyseFileRequest, analyseFileOutputChan chan<- OwnershipResult, analyseFileErrChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for req := range analyseFileInputChan {
		startTime := time.Now()
		ownershipResult := OwnershipResult{TotalLines: 0, authorLinesMap: make(map[string]int, 0)}
		ownershipResult.FilePath = req.filePath

		finfo, err := os.Stat(fmt.Sprintf("%s/%s", req.repoDir, req.filePath))
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Couldn't open file. file=%s. err=%s", req.filePath, err))
			break
		}
		if finfo.Size() > 80000 {
			logrus.Debugf("Ignoring file because it's too big. file=%s, size=%d", req.filePath, finfo.Size())
			continue
		}

		blameResult, err := utils.ExecGitBlame(req.repoDir, req.filePath, req.commitId)
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Error on git blame. file=%s. err=%s", req.filePath, err))
			break
		}

		ownershipResult.TotalFiles += 1
		for _, lineAuthor := range blameResult {
			if strings.Trim(lineAuthor.LineContents, " ") == "" {
				continue
			}
			ownershipResult.TotalLines += 1
			ownershipResult.authorLinesMap[lineAuthor.AuthorName] += 1
		}
		ownershipResult.blameTime = time.Since(startTime)
		analyseFileOutputChan <- ownershipResult
		// time.Sleep(1 * time.Second)
		// fmt.Printf("Time spent: %s\n", time.Since(startTime))
	}
}
