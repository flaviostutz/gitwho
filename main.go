package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/flaviostutz/gitwho/ownership"
	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
)

var (
	ownershipOpts ownership.OwnershipOptions
)

func main() {
	loglevel := logrus.DebugLevel

	// authorsFilesFlag := flag.NewFlagSet("", flag.ExitOnError)
	// authorsFilesFlag.StringVar(&authorsFilesOpts.authors, "authors", ".*", "Regex for filtering commits by author name. Defaults to '.*'")
	// authorsFilesFlag.StringVar(&authorsFilesOpts.branches, "branches", "main", "Regex for filtering commits by branch name. Defaults to 'main'")
	// authorsFilesFlag.StringVar(&authorsFilesOpts.branches, "files", ".*", "Regex for filtering which files paths to analyse. Defaults to '.*'")
	// authorsFilesFlag.StringVar(&authorsFilesOpts.from, "from", "", "Filter commits from this date. Defaults to the start of the repo.")
	// authorsFilesFlag.StringVar(&authorsFilesOpts.from, "to", "", "Filter commits until this date. Defaults to current date")

	ownershipFlag := flag.NewFlagSet("", flag.ExitOnError)
	ownershipFlag.StringVar(&ownershipOpts.Branch, "branch", "main", "Branch name to analyse. Defaults to 'main'")
	ownershipFlag.StringVar(&ownershipOpts.WhenStr, "when", "", "Date time to analyse. Defaults to 'now'")
	ownershipFlag.StringVar(&ownershipOpts.FilesRegex, "files", "", "Regex for selecting which file paths to include in analysis. Defaults to '.*'")

	logrus.SetLevel(loglevel)

	if len(os.Args) < 2 {
		fmt.Println("Expected 'authors', 'files' or 'ownership' command")
		os.Exit(1)
	}

	// load local dir
	repo, err := git.PlainOpen(".")
	if err != nil {
		fmt.Println("Cannot load git repo from current path. err=", err)
		os.Exit(2)
	}

	switch os.Args[1] {
	// case "authors":
	// 	authorsFilesFlag.Parse(os.Args[2:])
	// 	logrus.Debugf("Starting analysis of author changes")
	// case "files":
	// 	authorsFilesFlag.Parse(os.Args[2:])
	// 	logrus.Debugf("Starting analysis of file changes")
	case "ownership":
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

		logrus.Debugf("Starting analysis of code ownership")
		ownershipResults, err := ownership.AnalyseCodeOwnership(repo, ownershipOpts)
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
