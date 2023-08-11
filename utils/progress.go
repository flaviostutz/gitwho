package utils

import (
	"fmt"
	"strings"
	"time"
)

type ProgressInfo struct {
	TotalTasks         int
	TotalTasksKnown    bool
	CompletedTasks     int
	CompletedTotalTime time.Duration
	Message            string
}

func ShowProgress(progressChan <-chan ProgressInfo) {
	for pc := range progressChan {
		if pc.TotalTasks == 0 || pc.CompletedTasks == 0 {
			continue
		}

		perc := 100 * float32(pc.CompletedTasks) / float32(pc.TotalTasks)
		pending := ""
		if !pc.TotalTasksKnown {
			pending = "+"
		}
		avg := float64(pc.CompletedTotalTime.Milliseconds()) / float64(pc.CompletedTasks)

		filler := "                                        "
		fileName := ""
		i := strings.LastIndex(pc.Message, "/")
		if i != -1 {
			fileName = pc.Message[i+1:]
			if len(fileName) < len(filler) {
				fileName += filler[:len(filler)-len(fileName)]
			}
			if len(fileName) > 40 {
				fileName = fileName[:40]
			}
		}
		// fmt.Printf("%d%% (%d/%d%s) %s - %dms\n", int(perc), pc.CompletedTasks, pc.TotalTasks, pending, pc.Message, int(avg))
		fmt.Printf("%d%% (%d/%d%s) %dms %s \r \a", int(perc), pc.CompletedTasks, pc.TotalTasks, pending, int(avg), fileName)
	}
}
