package utils

import (
	"errors"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func GetBranchHash(repo *git.Repository, branchName string) (plumbing.Hash, error) {
	brefs, err := repo.Branches()
	if err != nil {
		return plumbing.Hash{}, err
	}

	var branchHash plumbing.Hash
	for {
		bref, err := brefs.Next()
		if err != nil {
			break
		}
		if bref.Name().Short() == branchName {
			branchHash = bref.Hash()
			break
		}
	}
	if branchHash.IsZero() {
		return plumbing.Hash{}, errors.New("Couldn't find branch " + branchName)
	}
	return branchHash, nil
}

func GetCommitHashForTime(repo *git.Repository, branchName string, when time.Time) (plumbing.Hash, error) {
	// find branch
	branchHash, err := GetBranchHash(repo, branchName)
	if err != nil {
		return plumbing.Hash{}, err
	}

	// walk through git log for this branch (DESC order)
	ci, err := repo.Log(&git.LogOptions{Order: git.LogOrderCommitterTime, From: branchHash})
	if err != nil {
		return plumbing.Hash{}, nil
	}
	prevCommit := plumbing.Hash{}
	for {
		c, err := ci.Next()
		if err != nil {
			break
		}
		if c.Author.When.Before(when) {
			prevCommit = c.Hash
			break
		}
	}
	if prevCommit.IsZero() {
		return plumbing.Hash{}, errors.New("Couldn't find a commit before date")
	}
	return prevCommit, nil
}
