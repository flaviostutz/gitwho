package ownership

import (
	"flag"
	"fmt"
	"os"

	"github.com/flaviostutz/gitwho/cli"
	"github.com/flaviostutz/gitwho/ownership"
	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
)

func RunOwnershipTimeseries(osArgs []string) {
	opts := ownership.OwnershipTimeseriesOptions{}
	cliOpts := cli.CliOpts{}

	flags := flag.NewFlagSet("ownership-timeseries", flag.ExitOnError)
	flags.StringVar(&opts.RepoDir, "repo", ".", "Repository path to analyse")
	flags.StringVar(&opts.Branch, "branch", "main", "Branch name to analyse")
	flags.StringVar(&opts.FilesRegex, "files", ".*", "Regex for selecting which file paths to include in analysis")
	flags.StringVar(&opts.FilesNotRegex, "files-not", "", "Regex for filtering out files from analysis")
	flags.StringVar(&opts.AuthorsRegex, "authors", ".*", "Regex for selecting which authors to include in analysis")
	flags.StringVar(&opts.AuthorsNotRegex, "authors-not", "", "Regex for filtering out authors from analysis")
	flags.StringVar(&opts.CacheFile, "cache-file", "", "If defined, stores results in a cache file that can be used in subsequent calls that uses the same parameters.")
	flags.IntVar(&opts.CacheTTLSeconds, "cache-ttl", 5184000, "Time in seconds for old items in cache file to be deleted. Defaults to 2 months")
	flags.StringVar(&opts.Since, "since", "3 months ago", "Starting date for historical analysis. Eg: '1 year ago'")
	flags.StringVar(&opts.Until, "until", "now", "Ending date for historical analysis. Eg: 'now'")
	flags.StringVar(&opts.Period, "period", "2 weeks", "Show ownership data each [period] in the range [since]-[until]. Eg.: '7 days', '1 month'")
	flags.IntVar(&opts.MinDuplicateLines, "min-dup-lines", 4, "Min number of similar lines in a row to be considered a duplicate")
	flags.StringVar(&cliOpts.Format, "format", "full", "Output format. 'full' (more details) or 'short' (lines per author)")
	flags.StringVar(&cliOpts.GoProfileFile, "profile-file", "", "Profile file to dump golang runtime data to")
	flags.BoolVar(&cliOpts.Verbose, "verbose", false, "Show verbose logs during processing")

	flags.Parse(osArgs[2:])

	progressChan := cli.SetupBasic(cliOpts)
	defer close(progressChan)

	_, err := utils.ExecCommitIdsInRange(opts.RepoDir, opts.Branch, "", "")
	if err != nil {
		fmt.Printf("Branch %s not found\n", opts.Branch)
		os.Exit(1)
	}

	logrus.Debugf("Starting analysis of code ownership")
	ownershipResults, err := ownership.AnalyseTimeseriesOwnership(opts, progressChan)
	if err != nil {
		fmt.Println("Failed to perform ownership-timeseries analysis. err=", err)
		os.Exit(2)
	}

	switch cliOpts.Format {
	case "full":
		str := FormatTimeseriesOwnershipResults(ownershipResults, true)
		fmt.Println(str)
	case "short":
		str := FormatTimeseriesOwnershipResults(ownershipResults, false)
		fmt.Println(str)
	case "graph":
		url := ServeOwnershipTimeseries(ownershipResults, opts)
		_, err := utils.ExecShellf("", "open %s", url)
		if err != nil {
			fmt.Printf("Couldn't open browser automatically. See results at %s\n", url)
		}
		fmt.Printf("\nServing graph at %s\n", url)
		select {}
	}
}
