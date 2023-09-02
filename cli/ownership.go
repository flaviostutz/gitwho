package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/flaviostutz/gitwho/ownership"
	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
)

func RunOwnership(osArgs []string) {
	opts := ownership.OwnershipOptions{}
	cliOpts := CliOpts{}
	when := ""
	flags := flag.NewFlagSet("ownership", flag.ExitOnError)
	flags.StringVar(&opts.RepoDir, "repo", ".", "Repository path to analyse")
	flags.StringVar(&opts.Branch, "branch", "main", "Branch name to analyse")
	flags.StringVar(&opts.FilesRegex, "files", ".*", "Regex for selecting which file paths to include in analysis")
	flags.StringVar(&opts.FilesNotRegex, "files-not", "", "Regex for filtering out files from analysis")
	flags.StringVar(&opts.AuthorsRegex, "authors", ".*", "Regex for selecting which authors to include in analysis")
	flags.StringVar(&opts.AuthorsNotRegex, "authors-not", "", "Regex for filtering out authors from analysis")
	flags.IntVar(&opts.MinDuplicateLines, "min-dup-lines", 4, "Min number of similar lines in a row to be considered a duplicate")
	flags.StringVar(&when, "when", "now", "Date to do analysis in repo")
	flags.StringVar(&cliOpts.Format, "format", "full", "Output format. 'full' (more details), 'short' (lines per author) or 'graph' (open browser)")
	flags.StringVar(&cliOpts.GoProfileFile, "profile-file", "", "Profile file to dump golang runtime data to")
	flags.BoolVar(&cliOpts.Verbose, "verbose", false, "Show verbose logs during processing")

	flags.Parse(osArgs[2:])

	progressChan := setupBasic(cliOpts)
	defer close(progressChan)

	commit, err := utils.ExecGetLastestCommit(opts.RepoDir, opts.Branch, "", when)
	if err != nil {
		fmt.Printf("Branch %s not found\n", opts.Branch)
		os.Exit(1)
	}
	opts.CommitId = commit.CommitId

	logrus.Debugf("Starting analysis of code ownership. commitId=%s", opts.CommitId)
	ownershipResult, err := ownership.AnalyseCodeOwnership(opts, progressChan)
	if err != nil {
		fmt.Println("Failed to perform ownership analysis. err=", err)
		os.Exit(2)
	}

	switch cliOpts.Format {
	case "full":
		output := ownership.FormatCodeOwnershipResults(ownershipResult, true)
		fmt.Println(output)

	case "short":
		output := ownership.FormatCodeOwnershipResults(ownershipResult, false)
		fmt.Println(output)

	case "graph":
		url := ownership.ServeOwnership(ownershipResult, opts)
		_, err := utils.ExecShellf("", "open %s", url)
		if err != nil {
			fmt.Printf("Couldn't open browser automatically. See results at %s\n", url)
		}
		fmt.Printf("Serving graph at %s\n", url)
		select {}
	}
}
