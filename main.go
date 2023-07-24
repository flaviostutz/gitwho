package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/flaviostutz/gitwho/ownership"
	"github.com/flaviostutz/gitwho/utils"
	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
)

var (
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
	// authorsFilesFlag := flag.NewFlagSet("", flag.ExitOnError)
	// authorsFilesFlag.StringVar(&authorsFilesOpts.authors, "authors", ".*", "Regex for filtering commits by author name. Defaults to '.*'")
	// authorsFilesFlag.StringVar(&authorsFilesOpts.branches, "branches", "main", "Regex for filtering commits by branch name. Defaults to 'main'")
	// authorsFilesFlag.StringVar(&authorsFilesOpts.branches, "files", ".*", "Regex for filtering which files paths to analyse. Defaults to '.*'")
	// authorsFilesFlag.StringVar(&authorsFilesOpts.from, "from", "", "Filter commits from this date. Defaults to the start of the repo.")
	// authorsFilesFlag.StringVar(&authorsFilesOpts.from, "to", "", "Filter commits until this date. Defaults to current date")

	ownershipFlag := flag.NewFlagSet("", flag.ExitOnError)
	ownershipFlag.StringVar(&ownershipOpts.Branch, "branch", "main", "Branch name to analyse. Defaults to 'main'")
	ownershipFlag.StringVar(&ownershipOpts.WhenStr, "when", "", "Date time to analyse. Defaults to 'now'")
	ownershipFlag.StringVar(&ownershipOpts.FilesRegex, "files", ".*", "Regex for selecting which file paths to include in analysis. Defaults to '.*'")
	ownershipFlag.StringVar(&ownershipOpts.RepoDir, "repo", ".", "Repository path to analyse. Defaults to current dir")
	ownershipFlag.BoolVar(&ownershipOpts.Verbose, "verbose", false, "Show verbose logs during processing. Defaults to false")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'authors', 'files' or 'ownership' command")
		os.Exit(1)
	}

	switch os.Args[1] {
	// case "authors":
	// 	authorsFilesFlag.Parse(os.Args[2:])
	// 	logrus.Debugf("Starting analysis of author changes")
	// case "files":
	// 	authorsFilesFlag.Parse(os.Args[2:])
	// 	logrus.Debugf("Starting analysis of file changes")
	case "ownership":
		// Start profiling
		f, err := os.Create("profile.pprof")
		if err != nil {
			fmt.Println(err)
			panic(5)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()

		ownershipFlag.Parse(os.Args[2:])

		// parse date
		if ownershipOpts.WhenStr == "" || ownershipOpts.WhenStr == "now" {
			ownershipOpts.WhenStr = time.Now().Format(time.RFC3339)
		}
		parsedTime, err := time.Parse(time.RFC3339, ownershipOpts.WhenStr)
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
		fmt.Println("Invalid command. Expected 'authors', 'files' or 'ownership'")
		os.Exit(1)
	}
}
