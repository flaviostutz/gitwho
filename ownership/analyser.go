package ownership

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

type OwnershipOptions struct {
	utils.BaseOptions
	When string
}

type AuthorLines struct {
	AuthorName           string
	AuthorMail           string
	OwnedLines           int
	OwnedLinesAgeDaysSum float64
}
type OwnershipResult struct {
	TotalFiles      int
	TotalLines      int
	LinesAgeDaysSum float64
	authorLinesMap  map[string]AuthorLines // temporary map used during processing
	AuthorsLines    []AuthorLines
	CommitId        string
	FilePath        string
	blameTime       time.Duration
	skippedFiles    int
}

type fileWorkerRequest struct {
	repoDir  string
	filePath string
	commitId string
}

func AnalyseCodeOwnership(opts OwnershipOptions, progressChan chan<- utils.ProgressInfo) (OwnershipResult, error) {
	result := OwnershipResult{TotalLines: 0, authorLinesMap: make(map[string]AuthorLines, 0), AuthorsLines: make([]AuthorLines, 0)}

	progressInfo := utils.ProgressInfo{}

	fre, err := regexp.Compile(opts.FilesRegex)
	if err != nil {
		return result, errors.New("file filter regex is invalid. err=" + err.Error())
	}

	freNot, err := regexp.Compile(opts.FilesNotRegex)
	if err != nil {
		return result, errors.New("files-not filter regex is invalid. err=" + err.Error())
	}

	logrus.Debugf("Analysing branch %s at %s", opts.Branch, opts.When)

	commitId, err := utils.ExecGetCommitAtDate(opts.RepoDir, opts.Branch, opts.When)
	if err != nil {
		return result, err
	}
	result.CommitId = commitId
	logrus.Debugf("Using commitId %s", commitId)

	// MAP REDUCE - analyse files in parallel goroutines
	// we need to start workers in the reverse order so that all the chain
	// is prepared when submitting tasks to avoid deadlocks
	nrWorkers := runtime.NumCPU() - 1
	// nrWorkers := 1
	logrus.Debugf("Preparing a pool of workers to process file analysis in parallel")
	fileWorkerInputChan := make(chan fileWorkerRequest, 5000)
	fileWorkerOutputChan := make(chan OwnershipResult, 5000)
	fileWorkerErrChan := make(chan error, nrWorkers)

	// REDUCE - summarise counters (STEP 3/3)
	var summaryWorkerWaitGroup sync.WaitGroup
	summaryWorkerWaitGroup.Add(1)
	go func() {
		defer summaryWorkerWaitGroup.Done()
		logrus.Debugf("Counting total lines owned per author")
		for fileResult := range fileWorkerOutputChan {
			result.TotalFiles += fileResult.TotalFiles
			result.TotalLines += fileResult.TotalLines
			result.LinesAgeDaysSum += fileResult.LinesAgeDaysSum
			for author := range fileResult.authorLinesMap {
				fileAuthorLines := fileResult.authorLinesMap[author]
				resultAuthorLines := result.authorLinesMap[author]
				resultAuthorLines.AuthorName = fileAuthorLines.AuthorName
				resultAuthorLines.AuthorMail = fileAuthorLines.AuthorMail
				resultAuthorLines.OwnedLines += fileAuthorLines.OwnedLines
				resultAuthorLines.OwnedLinesAgeDaysSum += fileAuthorLines.OwnedLinesAgeDaysSum
				result.authorLinesMap[author] = resultAuthorLines
			}
			progressInfo.CompletedTasks += 1 + fileResult.skippedFiles
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
			authorsLines = append(authorsLines, lines)
		}

		sort.Slice(authorsLines, func(i, j int) bool {
			return authorsLines[i].OwnedLines > authorsLines[j].OwnedLines
		})
		result.AuthorsLines = authorsLines
	}()

	// MAP - start analyser workers (STEP 2/3)
	var fileWorkersWaitGroup sync.WaitGroup
	for i := 0; i < nrWorkers; i++ {
		fileWorkersWaitGroup.Add(1)
		go fileWorker(fileWorkerInputChan, fileWorkerOutputChan, fileWorkerErrChan, &fileWorkersWaitGroup)
	}
	logrus.Debugf("Launched %d workers for analysis", nrWorkers)

	// MAP - submit tasks (STEP 1/3)
	go func() {
		logrus.Debugf("Scheduling files for analysis. filesRegex=%s", opts.FilesRegex)
		totalFiles := 0
		progressInfo.TotalTasksKnown = false
		files, err := utils.ExecListTree(opts.RepoDir, commitId)
		if err != nil {
			logrus.Errorf("Error getting commit tree. err=%s", err)
			panic(5)
		}

		for _, fileName := range files {
			if strings.Trim(fileName, " ") == "" || !fre.MatchString(fileName) || (opts.FilesNotRegex != "" && freNot.MatchString(fileName)) {
				// logrus.Debugf("Ignoring file %s", file.Name)
				continue
			}
			totalFiles += 1
			progressInfo.TotalTasks += 1
			fileWorkerInputChan <- fileWorkerRequest{repoDir: opts.RepoDir, filePath: fileName, commitId: commitId}
		}

		// finished publishing request messages
		logrus.Debugf("%d files scheduled for analysis", totalFiles)
		logrus.Debug("Task submission worker finished")
		close(fileWorkerInputChan)

		progressInfo.TotalTasksKnown = true
		if progressChan != nil {
			progressChan <- progressInfo
		}
	}()

	fileWorkersWaitGroup.Wait()
	logrus.Debug("Analysis workers finished")
	close(fileWorkerOutputChan)
	close(fileWorkerErrChan)

	for workerErr := range fileWorkerErrChan {
		logrus.Errorf("Error during analysis. err=%s", workerErr)
		panic(2)
	}

	summaryWorkerWaitGroup.Wait()
	logrus.Debug("Summary worker finished")

	// fmt.Printf("SUMMARY: %v\n", result)

	return result, nil
}

// this will be run by multiple goroutines
func fileWorker(fileWorkerInputChan <-chan fileWorkerRequest, fileWorkerOutputChan chan<- OwnershipResult, fileWorkerErrChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	skippedFiles := 0
	for req := range fileWorkerInputChan {
		startTime := time.Now()
		ownershipResult := OwnershipResult{TotalLines: 0, authorLinesMap: make(map[string]AuthorLines, 0)}
		ownershipResult.FilePath = req.filePath

		commitInfo, err := utils.ExecGitCommitInfo(req.repoDir, req.commitId)
		if err != nil {
			fileWorkerErrChan <- errors.New(fmt.Sprintf("Couldn't get commit info. commitId=%s; err=%s", req.commitId, err))
			break
		}

		fsize, err := utils.ExecTreeFileSize(req.repoDir, req.commitId, req.filePath)
		if err != nil {
			// can't get file size when the file was deleted by commit, so it's not present anymore
			// TODO get previous version of the file and count these lines as "changed" because they were deleted?
			skippedFiles++
			continue
		}
		if fsize > 80000 {
			logrus.Debugf("Ignoring file because it's too big. file=%s, size=%d", req.filePath, fsize)
			skippedFiles++
			continue
		}

		isBin, err := utils.ExecDiffIsBinary(req.repoDir, req.commitId, req.filePath)
		if err != nil {
			fileWorkerErrChan <- errors.New(fmt.Sprintf("Couldn't determine if file is binary. file=%s; commitId=%s; err=%s", req.filePath, req.commitId, err))
			break
		}
		if isBin {
			logrus.Debugf("Ignoring binary file. file=%s, commitId=%s", req.filePath, req.commitId)
			skippedFiles++
			continue
		}

		blameResult, err := utils.ExecGitBlame(req.repoDir, req.filePath, req.commitId)
		if err != nil {
			fileWorkerErrChan <- errors.New(fmt.Sprintf("Error on git blame. file=%s. err=%s", req.filePath, err))
			break
		}

		ownershipResult.TotalFiles += 1
		for _, lineAuthor := range blameResult {
			if strings.Trim(lineAuthor.LineContents, " ") == "" {
				continue
			}
			ownershipResult.TotalLines += 1
			authorLines := ownershipResult.authorLinesMap[lineAuthor.AuthorName]
			authorLines.AuthorName = lineAuthor.AuthorName
			authorLines.AuthorMail = lineAuthor.AuthorMail
			authorLines.OwnedLines += 1
			// fmt.Printf(">>>>%s %s %f %s\n", commitInfo.Date, lineAuthor.AuthorDate, commitInfo.Date.Sub(lineAuthor.AuthorDate).Hours(), commitInfo.Date.Sub(lineAuthor.AuthorDate))
			lineAge := (commitInfo.Date.Sub(lineAuthor.AuthorDate).Hours()) / float64(24)
			authorLines.OwnedLinesAgeDaysSum += lineAge
			ownershipResult.LinesAgeDaysSum += lineAge
			// fmt.Printf("]]]]%s\n", authorLines.OwnedLinesAgeSum)
			ownershipResult.authorLinesMap[lineAuthor.AuthorName] = authorLines
		}
		ownershipResult.blameTime = time.Since(startTime)
		ownershipResult.skippedFiles += skippedFiles
		skippedFiles = 0
		fileWorkerOutputChan <- ownershipResult
		// time.Sleep(1 * time.Second)
		// fmt.Printf("Time spent: %s\n", time.Since(startTime))
	}
}
