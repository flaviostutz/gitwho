package ownership

import (
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"sync"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	fsutil "github.com/go-git/go-billy/v5/util"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
)

type OwnershipOptions struct {
	Branch     string
	WhenStr    string
	When       time.Time
	FilesRegex string
}

type AuthorLines struct {
	Author     string
	OwnedLines uint
}

type OwnershipResult struct {
	TotalLines  uint
	AuthorLines map[string]int
}

type blameFileRequest struct {
	filePath string
}

func AnalyseCodeOwnership(repo *git.Repository, opts OwnershipOptions) (OwnershipResult, error) {
	result := OwnershipResult{}

	logrus.Debugf("Analysing branch %s at %s", opts.Branch, opts.When)

	chash, err := utils.GetCommitHashForTime(repo, opts.Branch, opts.When)
	if err != nil {
		return result, err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return result, err
	}
	logrus.Debugf("Checking out commit %s", chash.String())
	err = wt.Checkout(&git.CheckoutOptions{Hash: plumbing.NewHash(chash.String())})
	if err != nil {
		return result, err
	}

	// MAP - select and analyse files in parallel goroutines
	logrus.Debugf("Preparing a pool of workers to process file analysis in parallel")
	blameFileInputChan := make(chan blameFileRequest, 10)
	blameFileOutputChan := make(chan OwnershipResult, 10)

	// prepare worker pool to process file analysis in parallel
	var blameFileWaitGroup sync.WaitGroup
	for i := 0; i < 10; i++ {
		blameFileWaitGroup.Add(1)
		go blameFileWorker(blameFileInputChan, blameFileOutputChan, wt, &blameFileWaitGroup)
	}

	logrus.Debugf("Scheduling files for blame analysis. filesRegex=%s", opts.FilesRegex)
	fre, err := regexp.Compile(opts.FilesRegex)
	if err != nil {
		return result, errors.New("file filter regex is invalid. err=" + err.Error())
	}
	totalFiles := 0
	fsutil.Walk(wt.Filesystem, wt.Filesystem.Root(), func(path string, finfo fs.FileInfo, err error) error {
		if !finfo.IsDir() && finfo.Size() < 100000 && fre.MatchString(path) {
			totalFiles += 1
			// schedule file to be blamed by parallel workers
			blameFileInputChan <- blameFileRequest{filePath: path}
		}
		return nil
	})
	// finished publishing request messages
	close(blameFileInputChan)
	logrus.Debugf("%d files scheduled for analysis", totalFiles)

	// wait until all messages in blameFileInputChan chan are processed by workers
	blameFileWaitGroup.Wait()
	// finished publishing result messages
	close(blameFileOutputChan)

	// REDUCE - summarise counters
	logrus.Debugf("Counting total lines owned per author")
	summaryOwnership := OwnershipResult{TotalLines: 0, AuthorLines: make(map[string]int, 0)}
	for fileOwnership := range blameFileOutputChan {
		summaryOwnership.TotalLines += fileOwnership.TotalLines
		// FIXME do this for all authors
		summaryOwnership.AuthorLines["Flavio"] += fileOwnership.AuthorLines["Flavio"]
	}

	fmt.Printf("SUMMARY: %v\n", summaryOwnership)

	return result, nil
}

// this will be run by multiple goroutines
func blameFileWorker(blameFileInputChan chan blameFileRequest, blameFileOutputChan chan OwnershipResult, workingTree *git.Worktree, wg *sync.WaitGroup) {
	for req := range blameFileInputChan {
		fmt.Println(req.filePath)
		ownershipResult := OwnershipResult{TotalLines: 0, AuthorLines: make(map[string]int, 0)}
		//FIXME IMPLEMENT ANALYSIS
		ownershipResult.TotalLines = 10
		ownershipResult.AuthorLines["Flavio"] = 1
		blameFileOutputChan <- ownershipResult
	}
	wg.Done()
}
