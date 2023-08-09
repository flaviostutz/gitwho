package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecGetCommitAtDate(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	// should work for default master branch
	cid, err := ExecGetCommitAtDate(repoDir, "master", "now")
	assert.Nil(t, err)
	assert.NotEmpty(t, cid)

	// should fail for invalid branches
	cid, err = ExecGetCommitAtDate(repoDir, "invalid-branch", "now")
	assert.NotNil(t, err)
	assert.Empty(t, cid)
}

func TestExecListTree(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	cid, err := ExecGetCommitAtDate(repoDir, "master", "now")
	assert.Nil(t, err)
	assert.NotEmpty(t, cid)

	files, err := ExecListTree(repoDir, cid)
	assert.Nil(t, err)
	assert.NotEmpty(t, files)

	assert.Equal(t, 2, len(files))
}

func TestExecCommitDate(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	commitIds, err := ExecCommitsInRange(repoDir, "master", "1 month ago", "now")
	if err != nil {
		return
	}

	cinfo, err := ExecGitCommitInfo(repoDir, commitIds[0])
	assert.Nil(t, err)
	assert.False(t, cinfo.Date.IsZero())

	cinfo2, err := ExecGitCommitInfo(repoDir, commitIds[len(commitIds)-1])
	assert.Nil(t, err)
	assert.True(t, cinfo.Date.After(cinfo2.Date))
}

func TestExecDiffTree(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	cid, err := ExecGetCommitAtDate(repoDir, "master", "now")
	assert.Nil(t, err)
	assert.NotEmpty(t, cid)

	files, err := ExecDiffTree(repoDir, cid)
	assert.Nil(t, err)
	assert.NotEmpty(t, files)

	assert.Equal(t, 1, len(files))
}

func TestExecTreeFileSize(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	commitIds, err := ExecCommitsInRange(repoDir, "master", "1 month ago", "now")
	if err != nil {
		return
	}

	size1, err := ExecTreeFileSize(repoDir, commitIds[0], "file1")
	assert.Nil(t, err)
	assert.Equal(t, 3, size1)

	size2, err := ExecTreeFileSize(repoDir, commitIds[len(commitIds)-1], "file1")
	assert.Nil(t, err)
	assert.Equal(t, 1, size2)

	assert.NotEqual(t, size1, size2)
}

func TestExecDiffFileRevisions(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	commitIds, err := ExecCommitsInRange(repoDir, "master", "1 month ago", "now")
	if err != nil {
		return
	}

	de, err := ExecDiffFileRevisions(repoDir, "file1", commitIds[0], commitIds[len(commitIds)-1])
	assert.Nil(t, err)
	assert.Equal(t, OperationChange, de[0].Operation)
	assert.Equal(t, 1, de[0].DstLines[0].Number)
	assert.Equal(t, "a", de[0].DstLines[0].Text)
	assert.Equal(t, 2, de[0].SrcLines[1].Number)
	assert.Equal(t, "c", de[0].SrcLines[1].Text)
}

func TestExecCommitsInRange(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	cid, err := ExecCommitsInRange(repoDir, "master", "1 week ago", "now")
	assert.Nil(t, err)
	assert.NotEmpty(t, cid)
	assert.Equal(t, 5, len(cid))
}

func TestExecPreviousCommitIdForFile(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	cid, err := ExecGetCommitAtDate(repoDir, "master", "now")
	assert.Nil(t, err)
	assert.NotEmpty(t, cid)

	prevCid, err := ExecPreviousCommitIdForFile(repoDir, cid, "file1")
	assert.Nil(t, err)
	assert.NotEmpty(t, prevCid)
	assert.NotEqual(t, prevCid, cid)
}

func TestExecGitBlame(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	cid, err := ExecGetCommitAtDate(repoDir, "master", "now")
	assert.Nil(t, err)
	assert.NotEmpty(t, cid)

	lines, err := ExecGitBlame(repoDir, "file1", cid)
	assert.Nil(t, err)
	assert.NotEmpty(t, lines)
	assert.Equal(t, 2, len(lines))
	assert.Equal(t, "author2", lines[0].AuthorName)
	assert.Equal(t, "author1", lines[1].AuthorName)
}
