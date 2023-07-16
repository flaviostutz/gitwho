package ownership

import (
	"errors"
	"io/fs"
	"regexp"
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
	AuthorLines []AuthorLines
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

	// MAP - analyse each file in parallel
	logrus.Debugf("Preparing a pool of workers to process file analysis in parallel")
	// TODO prepare input and output channels

	logrus.Debugf("Scheduling workspace files for analysis. filesRegex=%s", opts.FilesRegex)
	fre, err := regexp.Compile(opts.FilesRegex)
	if err != nil {
		return result, errors.New("file filter regex is invalid. err=" + err.Error())
	}
	totalFiles := 0
	fsutil.Walk(wt.Filesystem, wt.Filesystem.Root(), func(path string, finfo fs.FileInfo, err error) error {
		if !finfo.IsDir() && finfo.Size() < 100000 && fre.MatchString(path) {
			totalFiles += 1
			// TODO publish to worker channel input
		}
		return nil
	})
	logrus.Debugf("%d files scheduled for analysis", totalFiles)

	// REDUCE - summarise counters
	logrus.Debugf("Counting lines owned per author")
	// TODO read and count from output channel

	return result, nil
}
