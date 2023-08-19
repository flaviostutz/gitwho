package changes

import (
	"testing"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestAnalyseWorkerFile1(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)

	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	analyseFileInputChan := make(chan fileWorkerRequest, 4)
	analyseFileOutputChan := make(chan ChangesFileResult, 4)

	commitIds, err := utils.ExecCommitsInRange(repoDir, "main", "1 month ago", "now")
	if err != nil {
		return
	}

	// submit commit1:file1 for analysis
	analyseFileInputChan <- fileWorkerRequest{repoDir: repoDir, commitId: commitIds[4], filePath: "file1"}
	analyseFileInputChan <- fileWorkerRequest{repoDir: repoDir, commitId: commitIds[3], filePath: "file1"}
	analyseFileInputChan <- fileWorkerRequest{repoDir: repoDir, commitId: commitIds[2], filePath: "file1"}
	analyseFileInputChan <- fileWorkerRequest{repoDir: repoDir, commitId: commitIds[1], filePath: "file1"}
	close(analyseFileInputChan)

	// execute analysis
	fileAnalysisWorker(analyseFileInputChan, analyseFileOutputChan, nil, nil)

	// require commit1 analysis
	// a1
	changes1 := <-analyseFileOutputChan
	require.Equal(t, "file1", changes1.FilePath)
	require.Equal(t, 1, changes1.TotalLines.New)
	require.Equal(t, 0, changes1.TotalLines.Changes)
	require.Equal(t, 0, changes1.TotalLines.ChurnOwn)
	require.Equal(t, 0, changes1.TotalLines.ChurnOther)
	require.Equal(t, 0, changes1.TotalLines.ChurnReceived)
	require.Equal(t, 0, changes1.TotalLines.RefactorOther)
	require.Equal(t, 0, changes1.TotalLines.RefactorOwn)
	require.Equal(t, 0, changes1.TotalLines.RefactorReceived)
	authorLines1, ok := changes1.authorLinesMap["author1###<author1@mail.com>"]
	require.True(t, ok)
	require.Equal(t, 1, authorLines1.LinesTouched.New)
	require.Equal(t, 0, authorLines1.LinesTouched.Changes)
	require.Equal(t, 0, authorLines1.LinesTouched.ChurnOwn)
	require.Equal(t, 0, authorLines1.LinesTouched.ChurnOther)
	require.Equal(t, 0, authorLines1.LinesTouched.ChurnReceived)
	require.Equal(t, 0, authorLines1.LinesTouched.RefactorOther)
	require.Equal(t, 0, authorLines1.LinesTouched.RefactorOwn)
	require.Equal(t, 0, authorLines1.LinesTouched.RefactorReceived)
	require.Equal(t, 1, len(authorLines1.filesTouchedMap))
	author1FilesMap, ok := authorLines1.filesTouchedMap["file1"]
	require.True(t, ok)
	require.Equal(t, "file1", author1FilesMap.Name)
	require.Equal(t, 1, author1FilesMap.Lines)

	// require commit2 analysis
	// a2
	changes2 := <-analyseFileOutputChan
	require.Equal(t, "file1", changes2.FilePath)
	require.Equal(t, 1, changes2.TotalLines.New)
	require.Equal(t, 1, changes2.TotalLines.Changes)
	require.Equal(t, 1, changes2.TotalLines.ChurnOther)
	require.Equal(t, 0, changes2.TotalLines.ChurnOwn)
	authorLines2, ok := changes2.authorLinesMap["author2###<author2@mail.com>"]
	require.True(t, ok)
	require.Equal(t, 1, authorLines2.LinesTouched.Changes)
	require.Equal(t, 1, authorLines2.LinesTouched.ChurnOther)
	require.Equal(t, 1, len(authorLines2.filesTouchedMap))
	author2FilesMap, ok := authorLines2.filesTouchedMap["file1"]
	require.True(t, ok)
	require.Equal(t, 2, author2FilesMap.Lines)

	// require commit3 analysis
	// a1
	changes3 := <-analyseFileOutputChan
	require.Equal(t, "file1", changes3.FilePath)
	require.Equal(t, 1, changes3.TotalLines.New)
	require.Equal(t, 1, changes3.TotalLines.Changes)
	require.Equal(t, 1, changes3.TotalLines.ChurnOther)
	require.Equal(t, 0, changes3.TotalLines.ChurnOwn)
	authorLines1, ok = changes3.authorLinesMap["author1###<author1@mail.com>"]
	require.True(t, ok)
	author1FilesMap, ok = authorLines1.filesTouchedMap["file1"]
	require.True(t, ok)
	require.Equal(t, 2, author1FilesMap.Lines)

	// require commit4 analysis
	// a1
	changes4 := <-analyseFileOutputChan
	require.Equal(t, "file1", changes4.FilePath)
	require.Equal(t, 0, changes4.TotalLines.New)
	require.Equal(t, 1, changes4.TotalLines.Changes)
	require.Equal(t, 1, changes4.TotalLines.ChurnOwn)
	require.Equal(t, 0, changes4.TotalLines.ChurnOther)
	authorLines4, ok := changes4.authorLinesMap["author1###<author1@mail.com>"]
	require.True(t, ok)
	require.Equal(t, 1, authorLines4.LinesTouched.ChurnOwn)
	require.Equal(t, 1, authorLines4.LinesTouched.Changes)
	require.Equal(t, 0, authorLines4.LinesTouched.ChurnOther)
	require.Equal(t, 1, len(authorLines4.filesTouchedMap))
	author1FilesMap, ok = authorLines4.filesTouchedMap["file1"]
	require.True(t, ok)
	require.Equal(t, 1, author1FilesMap.Lines)
}
