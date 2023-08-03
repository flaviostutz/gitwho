package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

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

	fromStr := ""
	toStr := ""
	profileFile := ""

	changesFlag := flag.NewFlagSet("changes", flag.ExitOnError)
	changesFlag.StringVar(&changesOpts.AuthorsRegex, "authors", ".*", "Regex for filtering changes by author name. Defaults to '.*'")
	changesFlag.StringVar(&changesOpts.Branch, "branch", "main", "Regex for filtering changes by branch name. Defaults to 'main'")
	changesFlag.StringVar(&changesOpts.FilesRegex, "files", ".*", "Regex for filtering which files paths to analyse. Defaults to '.*'")
	changesFlag.StringVar(&fromStr, "from", "", "Filter changes made from this date. Defaults to last 30 days")
	changesFlag.StringVar(&toStr, "to", "", "Filter changes made util this date. Defaults to current date")
	changesFlag.StringVar(&profileFile, "profile-file", "", "Profile file to dump golang runtime data to. Defaults to none")

	whenStr := ""
	ownershipFlag := flag.NewFlagSet("ownership", flag.ExitOnError)
	ownershipFlag.StringVar(&ownershipOpts.Branch, "branch", "main", "Branch name to analyse. Defaults to 'main'")
	ownershipFlag.StringVar(&whenStr, "when", "", "Date time to analyse. Defaults to 'now'")
	ownershipFlag.StringVar(&ownershipOpts.FilesRegex, "files", ".*", "Regex for selecting which file paths to include in analysis. Defaults to '.*'")
	ownershipFlag.StringVar(&ownershipOpts.RepoDir, "repo", ".", "Repository path to analyse. Defaults to current dir")
	ownershipFlag.BoolVar(&ownershipOpts.Verbose, "verbose", false, "Show verbose logs during processing. Defaults to false")
	ownershipFlag.StringVar(&profileFile, "profile-file", "", "Profile file to dump golang runtime data to. Defaults to none")

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

		// parse 'to' date
		if toStr == "" {
			toStr = time.Now().Format(time.RFC3339)
		}
		parsedToTime, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			fmt.Println("Invalid date used in 'to'")
			os.Exit(1)
		}
		changesOpts.To = parsedToTime

		// parse 'from' date
		if fromStr == "" {
			// defaults to one month before current time
			whenStr = parsedToTime.Add(-720 * time.Hour).Format(time.RFC3339)
		}
		parsedFromTime, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			fmt.Println("Invalid date used in 'from'")
			os.Exit(1)
		}
		changesOpts.From = parsedFromTime

		if parsedToTime.Before(parsedFromTime) {
			fmt.Println("Invalid dates: 'from' must be before 'to' date")
			os.Exit(1)
		}

		logrus.Debugf("Loading git repo at %s", changesOpts.RepoDir)
		repo, err := git.PlainOpen(changesOpts.RepoDir)
		if err != nil {
			fmt.Printf("Cannot load git repo at %s. err=%s", changesOpts.RepoDir, err)
			os.Exit(2)
		}

		logrus.Debugf("Starting analysis of code changes")
		progressChan := make(chan utils.ProgressInfo, 10)
		if ownershipOpts.Verbose {
			go utils.ShowProgress(progressChan)
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

		// parse date
		if whenStr == "" || whenStr == "now" {
			whenStr = time.Now().Format(time.RFC3339)
		}
		parsedTime, err := time.Parse(time.RFC3339, whenStr)
		if err != nil {
			fmt.Println("Invalid date used")
			os.Exit(1)
		}
		ownershipOpts.When = parsedTime

		logrus.Debugf("Loading git repo at %s", ownershipOpts.RepoDir)
		repo, err := git.PlainOpen(ownershipOpts.RepoDir)
		if err != nil {
			fmt.Printf("Cannot load git repo at %s. err=%s", ownershipOpts.RepoDir, err)
			os.Exit(2)
		}

		logrus.Debugf("Starting analysis of code ownership")
		progressChan := make(chan utils.ProgressInfo, 1)
		if ownershipOpts.Verbose {
			go utils.ShowProgress(progressChan)
		}

		ownershipResults, err := ownership.AnalyseCodeOwnership(repo, ownershipOpts, progressChan)
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
