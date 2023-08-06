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

	assert.Equal(t, len(files), 2)
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
	assert.Equal(t, len(lines), 2)
	assert.Equal(t, lines[0].AuthorName, "author2")
	assert.Equal(t, lines[1].AuthorName, "author1")
}
