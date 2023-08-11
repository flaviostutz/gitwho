package changes

import (
	"testing"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestAnalyseWorkerFile1(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	repoDir, err := utils.ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	analyseFileInputChan := make(chan analyseFileRequest, 4)
	analyseFileOutputChan := make(chan ChangesFileResult, 4)

	commitIds, err := utils.ExecCommitsInRange(repoDir, "main", "1 month ago", "now")
	if err != nil {
		return
	}

	// submit commit1:file1 for analysis
	analyseFileInputChan <- analyseFileRequest{repoDir: repoDir, commitId: commitIds[4], filePath: "file1"}
	analyseFileInputChan <- analyseFileRequest{repoDir: repoDir, commitId: commitIds[3], filePath: "file1"}
	analyseFileInputChan <- analyseFileRequest{repoDir: repoDir, commitId: commitIds[2], filePath: "file1"}
	analyseFileInputChan <- analyseFileRequest{repoDir: repoDir, commitId: commitIds[1], filePath: "file1"}
	close(analyseFileInputChan)

	// execute analysis
	analyseFileChangesWorker(analyseFileInputChan, analyseFileOutputChan, nil, nil)

	// assert commit1 analysis
	// a1
	changes1 := <-analyseFileOutputChan
	assert.Equal(t, "file1", changes1.FilePath)
	assert.Equal(t, 1, changes1.TotalLines.New)
	assert.Equal(t, 0, changes1.TotalLines.Changes)
	assert.Equal(t, 0, changes1.TotalLines.ChurnOwn)
	assert.Equal(t, 0, changes1.TotalLines.ChurnOther)
	authorLines1, ok := changes1.authorLinesMap["author1###<author1@mail.com>"]
	assert.True(t, ok)
	assert.Equal(t, 1, authorLines1.Lines.New)
	assert.Equal(t, 0, authorLines1.Lines.ChurnOther)

	// assert commit2 analysis
	// a2
	changes2 := <-analyseFileOutputChan
	assert.Equal(t, "file1", changes2.FilePath)
	assert.Equal(t, 1, changes2.TotalLines.New)
	assert.Equal(t, 1, changes2.TotalLines.Changes)
	assert.Equal(t, 1, changes2.TotalLines.ChurnOther)
	assert.Equal(t, 0, changes2.TotalLines.ChurnOwn)
	authorLines2, ok := changes2.authorLinesMap["author2###<author2@mail.com>"]
	assert.True(t, ok)
	assert.Equal(t, 1, authorLines2.Lines.Changes)
	assert.Equal(t, 1, authorLines2.Lines.ChurnOther)

	// assert commit3 analysis
	// a1
	changes3 := <-analyseFileOutputChan
	assert.Equal(t, "file1", changes3.FilePath)
	assert.Equal(t, 1, changes3.TotalLines.New)
	assert.Equal(t, 1, changes3.TotalLines.Changes)
	assert.Equal(t, 1, changes3.TotalLines.ChurnOther)
	assert.Equal(t, 0, changes3.TotalLines.ChurnOwn)

	// assert commit4 analysis
	// a1
	changes4 := <-analyseFileOutputChan
	assert.Equal(t, "file1", changes4.FilePath)
	assert.Equal(t, 0, changes4.TotalLines.New)
	assert.Equal(t, 1, changes4.TotalLines.Changes)
	assert.Equal(t, 1, changes4.TotalLines.ChurnOwn)
	assert.Equal(t, 0, changes4.TotalLines.ChurnOther)
	authorLines4, ok := changes4.authorLinesMap["author1###<author1@mail.com>"]
	assert.True(t, ok)
	assert.Equal(t, 1, authorLines4.Lines.ChurnOwn)
}
