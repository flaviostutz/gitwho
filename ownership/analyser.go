package ownership

import (
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	fsutil "github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
)

type OwnershipOptions struct {
	utils.BaseOptions
	When time.Time
}

type AuthorLines struct {
	Author     string
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
	filePath    string
	workingTree *git.Worktree
	commitObj   *object.Commit
}

func AnalyseCodeOwnership(repo *git.Repository, opts OwnershipOptions, progressChan chan<- utils.ProgressInfo) (OwnershipResult, error) {
	result := OwnershipResult{TotalLines: 0, authorLinesMap: make(map[string]int, 0), AuthorsLines: make([]AuthorLines, 0)}

	progressInfo := utils.ProgressInfo{}

	fre, err := regexp.Compile(opts.FilesRegex)
	if err != nil {
		return result, errors.New("file filter regex is invalid. err=" + err.Error())
	}

	logrus.Debugf("Analysing branch %s at %s", opts.Branch, opts.When)

	chash0, err := utils.GetCommitHashForTime(repo, opts.Branch, opts.When)
	if err != nil {
		return result, err
	}
	result.CommitId = chash0.String()

	commitObj, err := repo.CommitObject(chash0)
	if err != nil {
		return result, err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return result, err
	}
	logrus.Debugf("Checking out commit %s", chash0.String())
	err = wt.Checkout(&git.CheckoutOptions{
		Hash: plumbing.NewHash(chash0.String()),
		// SparseCheckoutDirectories: []string{"docs"},
	})
	if err != nil {
		return result, err
	}

	// MAP REDUCE - analyse files in parallel goroutines
	// we need to start workers in the reverse order so that all the chain
	// is prepared when submitting tasks to avoid deadlocks
	// nrWorkers := runtime.NumCPU() - 1
	nrWorkers := 1
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
			// FIXME remove later
			fmt.Printf("%s\n", fileResult.FilePath)
			progressInfo.CompletedTasks += 1
			progressInfo.Message = fmt.Sprintf("%s (%s)", fileResult.FilePath, fileResult.blameTime)
			if len(progressChan) < 1 {
				progressChan <- progressInfo
			}
		}

		logrus.Debugf("Sorting and preparing summary for each author")
		authorsLines := make([]AuthorLines, 0)
		for author := range result.authorLinesMap {
			lines := result.authorLinesMap[author]
			authorsLines = append(authorsLines, AuthorLines{Author: author, OwnedLines: lines})
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
		fsutil.Walk(wt.Filesystem, "/", func(path string, finfo fs.FileInfo, err error) error {
			// fmt.Printf("%s, %s, %s\n", path, finfo, err)
			if finfo == nil || finfo.IsDir() || finfo.Size() > 30000 || !fre.MatchString(path) || strings.Contains(path, "/.git/") {
				// logrus.Debugf("Ignoring file %s", finfo)
				return nil
			}
			totalFiles += 1

			// show progress
			progressInfo.TotalTasks += 1
			// if len(progressChan) < 1 {
			// 	progressChan <- progressInfo
			// }

			// schedule file to be blamed by parallel workers
			analyseFileInputChan <- analyseFileRequest{filePath: path, workingTree: wt, commitObj: commitObj}
			return nil
		})
		// finished publishing request messages
		logrus.Debugf("%d files scheduled for analysis", totalFiles)
		logrus.Debug("Task submission worker finished")
		close(analyseFileInputChan)

		progressInfo.TotalTasksKnown = true
		if len(progressChan) < 1 {
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
		blameResult, err := git.Blame(req.commitObj, strings.TrimLeft(req.filePath, "/"))
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Error on git blame. file=%s. err=%s", req.filePath, err))
			break
		}
		//TODO: IMPLEMENT IN GIT TO COMPARE SPEED
		ownershipResult.TotalFiles += 1
		for _, lineAuthor := range blameResult.Lines {
			if strings.Trim(lineAuthor.Text, " ") == "" {
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
