package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/flaviostutz/gitwho/changes"
	"github.com/flaviostutz/gitwho/ownership"
	"github.com/flaviostutz/gitwho/utils"
	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
)

var (
	changesOpts   changes.ChangesOptions
	ownershipOpts ownership.OwnershipOptions
)

type ProgressData struct {
	totalKnown     bool
	totalTasks     int
	completedTasks int
	info           string
}

func main() {

	logrus.SetLevel(logrus.DebugLevel)

	profileFile := ""
	verbose := false

	changesFlag := flag.NewFlagSet("changes", flag.ExitOnError)
	changesFlag.StringVar(&changesOpts.AuthorsRegex, "authors", ".*", "Regex for filtering changes by author name. Defaults to '.*'")
	changesFlag.StringVar(&changesOpts.Branch, "branch", "main", "Regex for filtering changes by branch name. Defaults to 'main'")
	changesFlag.StringVar(&changesOpts.FilesRegex, "files", ".*", "Regex for filtering which files paths to analyse. Defaults to '.*'")
	changesFlag.StringVar(&changesOpts.Since, "since", "", "Filter changes made from this date. Defaults to last 30 days")
	changesFlag.StringVar(&changesOpts.Until, "until", "", "Filter changes made util this date. Defaults to current date")
	changesFlag.StringVar(&profileFile, "profile-file", "", "Profile file to dump golang runtime data to. Defaults to none")
	changesFlag.BoolVar(&verbose, "verbose", true, "Show verbose logs during processing. Defaults to false")

	ownershipFlag := flag.NewFlagSet("ownership", flag.ExitOnError)
	ownershipFlag.StringVar(&ownershipOpts.Branch, "branch", "main", "Branch name to analyse. Defaults to 'main'")
	ownershipFlag.StringVar(&ownershipOpts.When, "when", "now", "Date time to analyse. Defaults to 'now'")
	ownershipFlag.StringVar(&ownershipOpts.FilesRegex, "files", ".*", "Regex for selecting which file paths to include in analysis. Defaults to '.*'")
	ownershipFlag.StringVar(&ownershipOpts.RepoDir, "repo", ".", "Repository path to analyse. Defaults to current dir")
	ownershipFlag.StringVar(&profileFile, "profile-file", "", "Profile file to dump golang runtime data to. Defaults to none")
	ownershipFlag.BoolVar(&verbose, "verbose", true, "Show verbose logs during processing. Defaults to false")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'gitwho changes' or 'gitwho ownership' command")
		os.Exit(1)
	}

	if profileFile != "" {
		// Start profiling
		f, err := os.Create(profileFile)
		if err != nil {
			logrus.Error(err)
			panic(5)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	switch os.Args[1] {
	case "changes":
		changesFlag.Parse(os.Args[2:])

		logrus.Debugf("Loading git repo at %s", changesOpts.RepoDir)
		repo, err := git.PlainOpen(changesOpts.RepoDir)
		if err != nil {
			fmt.Printf("Cannot load git repo at %s. err=%s", changesOpts.RepoDir, err)
			os.Exit(2)
		}

		logrus.Debugf("Starting analysis of code changes")
		progressChan := make(chan utils.ProgressInfo, 1)
		if verbose {
			go utils.ShowProgress(progressChan)
		} else {
			go func() {
				// simply empty progress
				for range progressChan {
				}
			}()
		}

		changesResults, err := changes.AnalyseChanges(repo, changesOpts, progressChan)
		close(progressChan)
		if err != nil {
			fmt.Println("Failed to perform changes analysis. err=", err)
			os.Exit(2)
		}
		output := changes.FormatTextResults(changesResults, changesOpts)
		fmt.Println(output)

	case "ownership":
		ownershipFlag.Parse(os.Args[2:])

		progressChan := make(chan utils.ProgressInfo, 1)
		if verbose {
			go utils.ShowProgress(progressChan)
		}

		logrus.Debugf("Starting analysis of code ownership")
		ownershipResults, err := ownership.AnalyseCodeOwnership(ownershipOpts, progressChan)
		close(progressChan)
		if err != nil {
			fmt.Println("Failed to perform ownership analysis. err=", err)
			os.Exit(2)
		}
		output := ownership.FormatTextResults(ownershipResults, ownershipOpts)
		fmt.Println(output)

	default:
		fmt.Println("Invalid command. Expected 'changes' or 'ownership'")
		os.Exit(1)
	}
}
