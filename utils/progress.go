package utils

import (
	"fmt"
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
	// fmt.Print("\033[s")
	for pc := range progressChan {
		if pc.TotalTasks == 0 || pc.CompletedTasks == 0 {
			continue
		}

		perc := 100 * float32(pc.CompletedTasks) / float32(pc.TotalTasks)
		pending := ""
		if !pc.TotalTasksKnown {
			pending = "+"
		}
		// fmt.Print("\033[u\033[K")
		// fmt.Printf("%d%% %s\n", int(perc), pc.Message)
		avg := float64(pc.CompletedTotalTime.Milliseconds()) / float64(pc.CompletedTasks)
		fmt.Printf("%d%% (%d/%d%s) %s - %dms\n", int(perc), pc.CompletedTasks, pc.TotalTasks, pending, pc.Message, int(avg))
	}
}
