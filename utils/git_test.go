package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/stretchr/testify/assert"
)

var (
	testRepo                *git.Repository
	testRepoFirstCommitHash string
	testRepoLastCommitHash  string
)

func GetTestRepo() *git.Repository {
	if testRepo != nil {
		return testRepo
	}

	fmt.Println("Creating test repo")
	fs := memfs.New()
	storer := memory.NewStorage()
	tRepo, err := git.Init(storer, fs)
	if err != nil {
		fmt.Println("Error initializing test repo")
		panic(1)
	}

	wt, err := tRepo.Worktree()

	// commit 1
	writeAddFile(fs, wt, "/file1", `a`)
	phash, _ := wt.Commit("commit 1", commitOptions("author1"))
	// fmt.Println("commit 1 = " + phash.String())
	testRepoFirstCommitHash = phash.String()
	time.Sleep(50 * time.Millisecond)

	// commit 2
	writeAddFile(fs, wt, "/file1", `a
b`)
	phash, _ = wt.Commit("commit 2", commitOptions("author2"))
	// fmt.Println("commit 2 = " + phash.String())
	time.Sleep(50 * time.Millisecond)

	// commit 3
	writeAddFile(fs, wt, "/file1", `a
d
c`)
	phash, _ = wt.Commit("commit 3", commitOptions("author1"))
	// fmt.Println("commit 3 = " + phash.String())
	time.Sleep(50 * time.Millisecond)

	// commit 4
	writeAddFile(fs, wt, "/file1", `a
c`)
	phash, _ = wt.Commit("commit 4", commitOptions("author1"))
	// fmt.Println("commit 4 = " + phash.String())
	testRepoLastCommitHash = phash.String()
	time.Sleep(50 * time.Millisecond)

	testRepo = tRepo
	return testRepo
}

func commitOptions(author string) *git.CommitOptions {
	return &git.CommitOptions{
		Author: &object.Signature{
			Name:  author,
			Email: author + "@mail.com",
			When:  time.Now(),
		},
	}
}

func writeAddFile(fs billy.Filesystem, wt *git.Worktree, file string, contents string) (err error) {
	file1, err := fs.Create(file)
	if err != nil {
		return err
	}
	file1.Write([]byte(contents))
	_, err1 := wt.Add(file)
	if err1 != nil {
		return err1
	}
	return nil
}

func TestGetBranchHash(t *testing.T) {
	repo := GetTestRepo()

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
	repo := GetTestRepo()

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
