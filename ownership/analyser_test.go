package ownership

import (
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
	require.Equal(t, 3, len(results.AuthorsLines))
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
