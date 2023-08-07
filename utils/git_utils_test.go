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

func TestExecPreviousCommitId(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	cid, err := ExecGetCommitAtDate(repoDir, "master", "now")
	assert.Nil(t, err)
	assert.NotEmpty(t, cid)

	prevCid, err := ExecPreviousCommitId(repoDir, cid)
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
