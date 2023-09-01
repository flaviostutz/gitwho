package cli

import (
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
)

type CliOpts struct {
	Verbose       bool
	GoProfileFile string
	Format        string
}

func setupBasic(cliOpts CliOpts) chan<- utils.ProgressInfo {
	if cliOpts.Format != "full" && cliOpts.Format != "short" && cliOpts.Format != "graph" {
		fmt.Println("'--format' should be (full|short|graph)")
		os.Exit(1)
	}

	logrus.SetLevel(logrus.WarnLevel)
	if cliOpts.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	utils.ExecCheckPrereqs()

	if cliOpts.GoProfileFile != "" {
		// Start profiling
		f, err := os.Create(cliOpts.GoProfileFile)
		if err != nil {
			logrus.Error(err)
			panic(5)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	progressChan := make(chan utils.ProgressInfo, 1)
	go utils.ShowProgress(progressChan)

	return progressChan
}
