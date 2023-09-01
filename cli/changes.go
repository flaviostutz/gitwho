package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/flaviostutz/gitwho/changes"
	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
)

func RunChanges(osArgs []string) {
	opts := changes.ChangesOptions{}
	cliOpts := CliOpts{}

	flags := flag.NewFlagSet("changes", flag.ExitOnError)
	flags.StringVar(&opts.RepoDir, "repo", ".", "Repository path to analyse")
	flags.StringVar(&opts.Branch, "branch", "main", "Branch name to analyse")
	flags.StringVar(&opts.FilesRegex, "files", ".*", "Regex for selecting which file paths to include in analysis")
	flags.StringVar(&opts.FilesNotRegex, "files-not", "", "Regex for filtering out files from analysis")
	flags.StringVar(&opts.AuthorsRegex, "authors", ".*", "Regex for selecting which authors to include in analysis")
	flags.StringVar(&opts.AuthorsNotRegex, "authors-not", "", "Regex for filtering out authors from analysis")
	flags.StringVar(&opts.Since, "since", "30 days ago", "Filter changes made from this date")
	flags.StringVar(&opts.Until, "until", "now", "Filter changes made util this date")
	flags.StringVar(&cliOpts.Format, "format", "full", "Output format. 'full' (more details) or 'short' (lines per author)")
	flags.StringVar(&cliOpts.GoProfileFile, "profile-file", "", "Profile file to dump golang runtime data to")
	flags.BoolVar(&cliOpts.Verbose, "verbose", false, "Show verbose logs during processing")

	flags.Parse(osArgs[2:])
	progressChan := setupBasic(cliOpts)
	defer close(progressChan)

	_, err := utils.ExecCommitsInRange(opts.RepoDir, opts.Branch, "", "")
	if err != nil {
		fmt.Printf("Branch %s not found\n", opts.Branch)
		os.Exit(1)
	}

	logrus.Debugf("Starting analysis of code changes")
	changesResults, err := changes.AnalyseChanges(opts, progressChan)
	if err != nil {
		fmt.Println("Failed to perform changes analysis. err=", err)
		os.Exit(2)
	}
	if cliOpts.Format == "short" {
		output := changes.FormatTopTextResults(changesResults)
		fmt.Println(output)
	} else {
		output := changes.FormatFullTextResults(changesResults)
		fmt.Println(output)
	}

	if changesResults.TotalCommits == 0 {
		fmt.Println("No changes found")
		os.Exit(3)
	}
}
