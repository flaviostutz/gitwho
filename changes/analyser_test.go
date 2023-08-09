package changes

import (
	"testing"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestAnalyseChangesNewFile2(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	logrus.SetLevel(logrus.DebugLevel)

	result, err := AnalyseChanges(ChangesOptions{
		BaseOptions: utils.BaseOptions{RepoDir: repoDir, Branch: "master", FilesRegex: "file2"},
	}, nil)

	// file2 was commited only one time with 5 lines of code

	assert.Nil(t, err)
	assert.Equal(t, 1, result.TotalCommits)
	assert.Equal(t, 1, result.TotalFiles)
	assert.Equal(t, LinesChanges{
		New:              5,
		Changes:          0,
		RefactorOwn:      0,
		RefactorOther:    0,
		RefactorReceived: 0,
		ChurnOwn:         0,
		ChurnOther:       0,
		ChurnReceived:    0,
		AgeSum:           time.Duration(0),
	}, result.TotalLines)

	assert.Equal(t, 1, len(result.AuthorsLines))
	assert.Equal(t, "author3", result.AuthorsLines[0].AuthorName)
	assert.Equal(t, LinesChanges{
		New:              5,
		Changes:          0,
		RefactorOwn:      0,
		RefactorOther:    0,
		RefactorReceived: 0,
		ChurnOwn:         0,
		ChurnOther:       0,
		ChurnReceived:    0,
		AgeSum:           time.Duration(0),
	}, result.AuthorsLines[0].Lines)
}

func TestAnalyseChangesFile1(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	logrus.SetLevel(logrus.DebugLevel)

	result, err := AnalyseChanges(ChangesOptions{
		BaseOptions: utils.BaseOptions{RepoDir: repoDir, Branch: "master", FilesRegex: "file1"},
	}, nil)

	// file1 was commited 4 times

	assert.Nil(t, err)
	assert.Equal(t, 4, result.TotalCommits)
	assert.Equal(t, 4, result.TotalFiles)
	assert.Equal(t, LinesChanges{
		New:              3,
		Changes:          2,
		RefactorOwn:      0,
		RefactorOther:    0,
		RefactorReceived: 0,
		ChurnOwn:         0,
		ChurnOther:       0,
		ChurnReceived:    0,
		AgeSum:           time.Duration(0),
	}, result.TotalLines)

	assert.Equal(t, 2, len(result.AuthorsLines))

	assert.Equal(t, "author1", result.AuthorsLines[0].AuthorName)
	assert.Equal(t, LinesChanges{
		New:              2,
		Changes:          2,
		RefactorOwn:      0,
		RefactorOther:    0,
		RefactorReceived: 0,
		ChurnOwn:         0,
		ChurnOther:       0,
		ChurnReceived:    0,
		AgeSum:           time.Duration(0),
	}, result.AuthorsLines[0].Lines)

	assert.Equal(t, "author2", result.AuthorsLines[1].AuthorName)
	assert.Equal(t, LinesChanges{
		New:              1,
		Changes:          0,
		RefactorOwn:      0,
		RefactorOther:    0,
		RefactorReceived: 0,
		ChurnOwn:         0,
		ChurnOther:       0,
		ChurnReceived:    0,
		AgeSum:           time.Duration(0),
	}, result.AuthorsLines[0].Lines)

}
