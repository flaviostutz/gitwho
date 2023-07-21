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
	Branch     string
	WhenStr    string
	When       time.Time
	FilesRegex string
	RepoDir    string
}

type AuthorLines struct {
	Author     string
	OwnedLines int
}
type OwnershipResult struct {
	TotalLines     int
	authorLinesMap map[string]int // temporary map used during processing
	AuthorsLines   []AuthorLines
	CommitId       string
}

type blameFileRequest struct {
	filePath    string
	workingTree *git.Worktree
	commitObj   *object.Commit
}

func AnalyseCodeOwnership(repo *git.Repository, opts OwnershipOptions) (OwnershipResult, error) {
	result := OwnershipResult{TotalLines: 0, authorLinesMap: make(map[string]int, 0), AuthorsLines: make([]AuthorLines, 0)}

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
	err = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(chash0.String())})
	if err != nil {
		return result, err
	}

	// MAP - analyse files in parallel goroutines
	BLAME_WORKERS := 5
	logrus.Debugf("Preparing a pool of workers to process file analysis in parallel")
	blameFileInputChan := make(chan blameFileRequest, 10)
	blameFileOutputChan := make(chan OwnershipResult, 10)
	blameFileErrChan := make(chan error, BLAME_WORKERS)

	// MAP - prepare worker pool
	var blameWorkersWaitGroup sync.WaitGroup
	for i := 0; i < BLAME_WORKERS; i++ {
		blameWorkersWaitGroup.Add(1)
		go blameFileWorker(blameFileInputChan, blameFileOutputChan, blameFileErrChan, &blameWorkersWaitGroup)
	}
	logrus.Debugf("Launched %d workers for blame analysis", BLAME_WORKERS)

	fmt.Printf(">>>>%s\n", opts.FilesRegex)
	fre, err := regexp.Compile(opts.FilesRegex)
	if err != nil {
		return result, errors.New("file filter regex is invalid. err=" + err.Error())
	}

	// MAP - submit tasks
	go func() {
		blameWorkersWaitGroup.Add(1)
		defer blameWorkersWaitGroup.Done()
		defer close(blameFileInputChan)
		logrus.Debugf("Scheduling files for blame analysis. filesRegex=%s", opts.FilesRegex)
		totalFiles := 0
		fsutil.Walk(wt.Filesystem, wt.Filesystem.Root(), func(path string, finfo fs.FileInfo, err error) error {
			fmt.Printf("%s, %s, %s\n", path, finfo, err)
			if finfo == nil || finfo.IsDir() || finfo.Size() > 30000 || !fre.MatchString(path) {
				logrus.Debugf("Ignoring file %s", finfo)
				return nil
			}
			totalFiles += 1
			// schedule file to be blamed by parallel workers
			blameFileInputChan <- blameFileRequest{filePath: path, workingTree: wt, commitObj: commitObj}
			return nil
		})
		// finished publishing request messages
		logrus.Debugf("%d files scheduled for analysis", totalFiles)
	}()

	// REDUCE - summarise counters
	var blameReduceWaitGroup sync.WaitGroup
	go func() {
		blameReduceWaitGroup.Add(1)
		defer blameReduceWaitGroup.Done()
		logrus.Debugf("Counting total lines owned per author")
		for fileResult := range blameFileOutputChan {
			result.TotalLines += fileResult.TotalLines
			for author := range fileResult.authorLinesMap {
				authorLines := fileResult.authorLinesMap[author]
				result.authorLinesMap[author] = authorLines + result.authorLinesMap[author]
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

	// wait until all messages in blameFileInputChan chan are processed by map workers and by reducer
	logrus.Debugf("Waiting processing to finish")

	blameWorkersWaitGroup.Wait()
	logrus.Debug("Blame workers finished")
	close(blameFileOutputChan)
	close(blameFileErrChan)

	for workerErr := range blameFileErrChan {
		logrus.Errorf("Error on blame analysis. err=%s", workerErr)
		panic(2)
	}

	blameReduceWaitGroup.Wait()
	logrus.Debug("Summarizer finished")

	// fmt.Printf("SUMMARY: %v\n", result)

	return result, nil
}

// this will be run by multiple goroutines
func blameFileWorker(blameFileInputChan <-chan blameFileRequest, blameFileOutputChan chan<- OwnershipResult, blameFileErrChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for req := range blameFileInputChan {
		ownershipResult := OwnershipResult{TotalLines: 0, authorLinesMap: make(map[string]int, 0)}
		blameResult, err := git.Blame(req.commitObj, strings.TrimLeft(req.filePath, "/"))
		if err != nil {
			blameFileErrChan <- err
			break
		}
		for _, lineAuthor := range blameResult.Lines {
			if strings.Trim(lineAuthor.Text, " ") == "" {
				continue
			}
			ownershipResult.TotalLines += 1
			ownershipResult.authorLinesMap[lineAuthor.Author] += 1
		}
		blameFileOutputChan <- ownershipResult
	}
}
