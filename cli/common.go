package cli

import (
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"runtime/pprof"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/sirupsen/logrus"
)

type CliOpts struct {
	Verbose       bool
	GoProfileFile string
	Format        string
}

func SetupBasic(cliOpts CliOpts) chan<- utils.ProgressInfo {
	if cliOpts.Format != "full" && cliOpts.Format != "short" && cliOpts.Format != "graph" && cliOpts.Format != "csv" {
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

func ServeGraphPage(page *components.Page, contents string) (string, *http.Server) {
	port := rand.Intn(20000) + 20000
	bindURL := fmt.Sprintf(":%d", port)

	srv := &http.Server{
		Addr: bindURL,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logrus.Debugf("Render page at %s", bindURL)
			page.Render(w)
			w.Write([]byte(contents))
		}),
	}

	go srv.ListenAndServe()

	return fmt.Sprintf("http://localhost%s", bindURL), srv
}
