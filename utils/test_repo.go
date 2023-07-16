package utils

import (
	"fmt"
	"time"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

var (
	testRepo                *git.Repository
	testRepoFirstCommitHash string
	testRepoLastCommitHash  string
)

func GetTestOwnershipRepo() (*git.Repository, error) {
	if testRepo != nil {
		return testRepo, nil
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
	if err != nil {
		return nil, err
	}

	// ROOT DIR

	// commit 1
	writeAddFile(fs, wt, "file1", `a`)
	phash, _ := wt.Commit("commit 1", commitOptions("author1"))
	// fmt.Println("commit 1 = " + phash.String())
	testRepoFirstCommitHash = phash.String()
	time.Sleep(50 * time.Millisecond)

	// commit 2
	writeAddFile(fs, wt, "file1", `a
b`)
	phash, _ = wt.Commit("commit 2", commitOptions("author2"))
	// fmt.Println("commit 2 = " + phash.String())
	time.Sleep(50 * time.Millisecond)

	// commit 3
	writeAddFile(fs, wt, "file1", `a
d
c`)
	phash, _ = wt.Commit("commit 3", commitOptions("author1"))
	// fmt.Println("commit 3 = " + phash.String())
	time.Sleep(50 * time.Millisecond)

	// commit 4
	writeAddFile(fs, wt, "file1", `a
c`)
	phash, _ = wt.Commit("commit 4", commitOptions("author1"))
	// fmt.Println("commit 4 = " + phash.String())
	testRepoLastCommitHash = phash.String()
	time.Sleep(50 * time.Millisecond)

	// DIR /dir1
	// commit 5
	writeAddFile(fs, wt, "dir1/dir1.1/file2", `a
b
c
d
e`)
	phash, _ = wt.Commit("commit 5", commitOptions("author3"))
	fmt.Println("commit 5 = " + phash.String())
	testRepoLastCommitHash = phash.String()
	time.Sleep(50 * time.Millisecond)

	testRepo = tRepo
	return testRepo, nil
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

func commitOptions(author string) *git.CommitOptions {
	return &git.CommitOptions{
		Author: &object.Signature{
			Name:  author,
			Email: author + "@mail.com",
			When:  time.Now(),
		},
	}
}
