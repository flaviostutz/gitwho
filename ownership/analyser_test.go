package ownership

import (
	"testing"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestAnalyseCodeOwnershipAllFiles(t *testing.T) {
	// assert.InDeltaf(t, float64(0), v, 0.01, "")
	repoDir, err := utils.ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "master",
		},
		When: "now",
	}, nil)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, 7, results.TotalLines)
	assert.Equal(t, 3, len(results.AuthorsLines))
}

func TestAnalyseCodeOwnershipRegexFiles(t *testing.T) {
	// assert.InDeltaf(t, float64(0), v, 0.01, "")
	repo, err := utils.ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir:    repo,
			Branch:     "main",
			FilesRegex: "/dir1.1/",
		},
		When: "now",
	}, nil)
	assert.Nil(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, 5, results.TotalLines)
	assert.Equal(t, 1, len(results.AuthorsLines))
}
