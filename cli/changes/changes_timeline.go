package changes

import (
	"flag"
	"fmt"
	"os"

	"github.com/flaviostutz/gitwho/changes"
	"github.com/flaviostutz/gitwho/cli"
	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
)

func RunChangesTimeseries(osArgs []string) {
	opts := changes.ChangesTimeseriesOptions{}
	cliOpts := cli.CliOpts{}

	flags := flag.NewFlagSet("changes-timeseries", flag.ExitOnError)
	flags.StringVar(&opts.RepoDir, "repo", ".", "Repository path to analyse")
	flags.StringVar(&opts.Branch, "branch", "main", "Branch name to analyse")
	flags.StringVar(&opts.FilesRegex, "files", ".*", "Regex for selecting which file paths to include in analysis")
	flags.StringVar(&opts.FilesNotRegex, "files-not", "", "Regex for filtering out files from analysis")
	flags.StringVar(&opts.AuthorsRegex, "authors", ".*", "Regex for selecting which authors to include in analysis")
	flags.StringVar(&opts.AuthorsNotRegex, "authors-not", "", "Regex for filtering out authors from analysis")
	flags.StringVar(&opts.CacheFile, "cache-file", "", "If defined, stores results in a cache file that can be used in subsequent calls that uses the same parameters.")
	flags.IntVar(&opts.CacheTTLSeconds, "cache-ttl", 5184000, "Time in seconds for old items in cache file to be deleted. Defaults to 2 months")
	flags.StringVar(&opts.Since, "since", "90 days ago", "Filter changes made from this date")
	flags.StringVar(&opts.Until, "until", "now", "Filter changes made util this date")
	flags.StringVar(&opts.Period, "period", "30 days ago", "Show changes data each [period] in the range [since]-[until]. Eg.: '7 days', '1 month'")
	flags.StringVar(&cliOpts.Format, "format", "full", "Output format. 'full' (more details), 'short' (lines per author) or 'graph' (open browser)")
	flags.StringVar(&cliOpts.GoProfileFile, "profile-file", "", "Profile file to dump golang runtime data to")
	flags.BoolVar(&cliOpts.Verbose, "verbose", false, "Show verbose logs during processing")

	flags.Parse(osArgs[2:])
	progressChan := cli.SetupBasic(cliOpts)
	defer close(progressChan)

	_, err := utils.ExecCommitIdsInDateRange(opts.RepoDir, opts.Branch, "", "")
	if err != nil {
		fmt.Printf("Branch %s not found\n", opts.Branch)
		os.Exit(1)
	}

	logrus.Debugf("Starting analysis of code changes")
	changesResults, err := changes.AnalyseTimeseriesChanges(opts, progressChan)
	if err != nil {
		fmt.Println("Failed to perform changes analysis. err=", err)
		os.Exit(2)
	}

	if len(changesResults) == 0 {
		fmt.Println("No changes found")
		os.Exit(3)
	}

	switch cliOpts.Format {
	case "full":
		output := FormatTimeseriesChangesResults(changesResults, true)
		fmt.Println(output)

	case "short":
		output := FormatTimeseriesChangesResults(changesResults, false)
		fmt.Println(output)

	case "graph":
		url := ServeChangesTimeseries(changesResults, opts)
		_, err := utils.ExecShellf("", "open %s", url)
		if err != nil {
			fmt.Printf("Couldn't open browser automatically. See results at %s\n", url)
		}
		fmt.Printf("\nServing graph at %s\n", url)
		select {}
	}
}
