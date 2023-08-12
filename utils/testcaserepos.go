package utils

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	ownershipRepoDir        *string
	testRepoFirstCommitHash string
	testRepoLastCommitHash  string
)

func ResolveTestOwnershipRepo() (string, error) {
	if ownershipRepoDir != nil {
		return *ownershipRepoDir, nil
	}

	curDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	testCasesDir := curDir + "/.testcaserepos"
	repoDir := testCasesDir + "/ownership"
	fmt.Printf("repoDir=%s\n", repoDir)

	// remove repo if exists
	_, err = ExecShellf("", "rm -rf %s", repoDir)
	if err != nil {
		return "", err
	}

	// create base dir for testcases
	ExecShellf("", "mkdir -p %s", testCasesDir)

	fmt.Println("Creating test repo")
	_, err = ExecShellf(testCasesDir, "git init ownership --initial-branch main")
	if err != nil {
		return "", err
	}

	// DON'T CHANGE THE REPO CONTENTS
	// there are complex unit tests that depends exactly on how it is

	// commit 1
	err = writeAddFile(repoDir, "file1", `a`)
	if err != nil {
		return "", err
	}
	testRepoFirstCommitHash, err = createCommit(repoDir, "commit 1", "author1")
	if err != nil {
		return "", err
	}

	// commit 2
	writeAddFile(repoDir, "file1", `a
b`)
	createCommit(repoDir, "commit 2", "author2")

	// commit 3
	writeAddFile(repoDir, "file1", `a
d
c`)
	createCommit(repoDir, "commit 3", "author1")

	// commit 4
	writeAddFile(repoDir, "file1", `a
c`)
	time.Sleep(1100 * time.Millisecond)
	testRepoLastCommitHash, _ = createCommit(repoDir, "commit 4", "author1")

	// DIR /dir1
	// commit 5
	writeAddFile(repoDir, "dir1/dir1.1/file2", `a
b
c
d
e`)
	testRepoLastCommitHash, _ = createCommit(repoDir, "commit 5", "author3")

	ownershipRepoDir = &repoDir
	return repoDir, nil
}

func writeAddFile(repoDir string, filePath string, contents string) error {
	fileDir := repoDir
	i := strings.LastIndex(filePath, "/")
	if i != -1 {
		fileDir = fmt.Sprintf("%s/%s", repoDir, filePath[:i])
	}
	err := os.MkdirAll(fileDir, os.ModePerm)
	if err != nil {
		return err
	}

	d1 := []byte(contents)
	err = os.WriteFile(fmt.Sprintf("%s/%s", repoDir, filePath), d1, 0644)
	if err != nil {
		return err
	}

	_, err = ExecShellf(repoDir, "git add %s", filePath)
	if err != nil {
		return err
	}

	return nil
}

func createCommit(repoDir string, comment string, author string) (string, error) {
	cmdResult, err := ExecShellf(repoDir, "git commit -m \"%s\" --author=\"%s <%s@mail.com>\"", comment, author, author)
	if err != nil {
		return "", err
	}
	re := regexp.MustCompile("\\s(.*)\\]")
	matches := re.FindStringSubmatch(cmdResult)
	if matches == nil {
		return "", fmt.Errorf("Couldn't find commit id in the result of commit. result=%s", cmdResult)
	}
	return matches[0], nil
}
