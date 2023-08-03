package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetBranchHash(t *testing.T) {
	repo, err := GetTestOwnershipRepo()
	assert.Nil(t, err)
	if err != nil {
		return
	}

	// should work for default master branch
	cid, err := GetBranchHash(repo, "master")
	assert.Nil(t, err)
	assert.False(t, cid.IsZero())

	// should fail for invalid branches
	cid, err = GetBranchHash(repo, "my-invalid-branch")
	assert.NotNil(t, err)
	assert.True(t, cid.IsZero())
}

func TestGetCommitHashForDate(t *testing.T) {
	repo, err := GetTestOwnershipRepo()
	assert.Nil(t, err)

	// should work on default master branch for a time after commits
	cid, err := GetCommitHashForTime(repo, "master", time.Now())
	assert.Nil(t, err)
	assert.Equal(t, testRepoLastCommitHash, cid.String())

	// should fail on default master branch for a time before commits
	cid, err = GetCommitHashForTime(repo, "master", time.Now().AddDate(0, 0, -1))
	assert.NotNil(t, err)

	// should fail for invalid branch
	cid, err = GetCommitHashForTime(repo, "invalid-branch", time.Now())
	assert.NotNil(t, err)
	assert.True(t, cid.IsZero())
}
