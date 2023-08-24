package ownership

import (
	"fmt"
	"testing"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestAnalyseCodeOwnershipAllFiles(t *testing.T) {
	// require.InDeltaf(t, float64(0), v, 0.01, "")
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		When: "now",
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}
	require.Equal(t, 7, results.TotalLines)
	require.Equal(t, 2, results.TotalFiles)
	require.Equal(t, 0, results.TotalLinesDuplicated)
	require.Equal(t, 3, len(results.AuthorsLines))

	sumLines := 0
	for _, al := range results.AuthorsLines {
		sumLines += al.OwnedLinesTotal
	}
	require.Equal(t, results.TotalLines, sumLines)
}

func TestAnalyseCodeOwnershipCheckSums(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipDuplicatesRepo()
	require.Nil(t, err)
	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		When: "now",
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}

	sumLines := 0
	sumDup := 0
	sumDupOrigOthers := 0
	for _, al := range results.AuthorsLines {
		sumLines += al.OwnedLinesTotal
		sumDup += al.OwnedLinesDuplicate
		sumDupOrigOthers += al.OwnedLinesDuplicateOriginalOthers
	}
	fmt.Printf("%d\n", results.TotalLinesDuplicated)
	require.Equal(t, results.TotalLines, sumLines)
	require.Equal(t, results.TotalLinesDuplicated, sumDup)
	require.Equal(t, results.TotalLinesDuplicated, sumDupOrigOthers)
}

func TestAnalyseCodeOwnershipRegexFiles(t *testing.T) {
	// require.InDeltaf(t, float64(0), v, 0.01, "")
	repo, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir:    repo,
			Branch:     "main",
			FilesRegex: "/dir1.1/",
		},
		When: "now",
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}
	require.Equal(t, 5, results.TotalLines)
	require.Equal(t, 1, len(results.AuthorsLines))
}

func TestAnalyseCodeOwnershipRegexNotFiles(t *testing.T) {
	// require.InDeltaf(t, float64(0), v, 0.01, "")
	repo, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir:       repo,
			Branch:        "main",
			FilesRegex:    "/dir1.1/",
			FilesNotRegex: "/dir1.1/",
		},
		When: "now",
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}
	require.Equal(t, 0, results.TotalLines)
	require.Equal(t, 0, len(results.AuthorsLines))
}
