package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/flaviostutz/gitwho/changes"
	"github.com/flaviostutz/gitwho/ownership"
	"github.com/flaviostutz/gitwho/utils"
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

	utils.ExecCheckPrereqs()

	format := "full"
	profileFile := ""
	verbose := false

	changesFlag := flag.NewFlagSet("changes", flag.ExitOnError)
	// changesFlag.StringVar(&changesOpts.AuthorsRegex, "authors", ".*", "Regex for filtering changes by author name")
	changesFlag.StringVar(&changesOpts.RepoDir, "repo", ".", "Repository path to analyse")
	changesFlag.StringVar(&changesOpts.Branch, "branch", "main", "Regex for filtering changes by branch name")
	changesFlag.StringVar(&changesOpts.FilesRegex, "files", ".*", "Regex for filtering which files paths to analyse")
	changesFlag.StringVar(&changesOpts.FilesNotRegex, "files-not", "", "Regex for filtering out files from analysis")
	changesFlag.StringVar(&changesOpts.Since, "since", "30 days ago", "Filter changes made from this date")
	changesFlag.StringVar(&changesOpts.Until, "until", "now", "Filter changes made util this date")
	changesFlag.StringVar(&profileFile, "profile-file", "", "Profile file to dump golang runtime data to")
	changesFlag.StringVar(&format, "format", "full", "Output format. 'full' (all authors with details) or 'short' (top authors by change type)")
	changesFlag.BoolVar(&verbose, "verbose", true, "Show verbose logs during processing")

	ownershipFlag := flag.NewFlagSet("ownership", flag.ExitOnError)
	ownershipFlag.StringVar(&ownershipOpts.RepoDir, "repo", ".", "Repository path to analyse")
	ownershipFlag.StringVar(&ownershipOpts.Branch, "branch", "main", "Branch name to analyse")
	ownershipFlag.StringVar(&ownershipOpts.When, "when", "now", "Date time to analyse")
	ownershipFlag.StringVar(&ownershipOpts.FilesRegex, "files", ".*", "Regex for selecting which file paths to include in analysis")
	ownershipFlag.StringVar(&changesOpts.FilesNotRegex, "files-not", "", "Regex for filtering out files from analysis")
	ownershipFlag.StringVar(&profileFile, "profile-file", "", "Profile file to dump golang runtime data to")
	ownershipFlag.BoolVar(&verbose, "verbose", true, "Show verbose logs during processing")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'gitwho changes' or 'gitwho ownership' command")
		os.Exit(1)
	}

	if format != "full" && format != "short" {
		fmt.Println("'--format' should be 'full or 'short''")
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

		_, err := utils.ExecCommitsInRange(changesOpts.RepoDir, changesOpts.Branch, "", "")
		if err != nil {
			fmt.Printf("Branch not found\n")
			os.Exit(1)
		}

		logrus.Debugf("Starting analysis of code changes")
		progressChan := make(chan utils.ProgressInfo, 1)
		if verbose {
			go utils.ShowProgress(progressChan)
		}

		changesResults, err := changes.AnalyseChanges(changesOpts, progressChan)
		close(progressChan)
		if err != nil {
			fmt.Println("Failed to perform changes analysis. err=", err)
			os.Exit(2)
		}
		if format == "short" {
			output := changes.FormatTopTextResults(changesResults, changesOpts)
			fmt.Println(output)
		} else {
			output := changes.FormatFullTextResults(changesResults, changesOpts)
			fmt.Println(output)
		}

		if changesResults.TotalCommits == 0 {
			os.Exit(3)
		}

	case "ownership":
		ownershipFlag.Parse(os.Args[2:])

		_, err := utils.ExecCommitsInRange(ownershipOpts.RepoDir, ownershipOpts.Branch, "", "")
		if err != nil {
			fmt.Printf("Branch not found\n")
			os.Exit(1)
		}

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
