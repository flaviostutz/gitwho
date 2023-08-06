package utils

import "fmt"

type ProgressInfo struct {
	TotalTasks      int
	TotalTasksKnown bool
	CompletedTasks  int
	Message         string
}

func ShowProgress(progressChan <-chan ProgressInfo) {
	// fmt.Print("\033[s")
	// fmt.Println(11111111)
	for pc := range progressChan {
		perc := 100 * float32(pc.CompletedTasks) / float32(pc.TotalTasks)
		pending := ""
		if !pc.TotalTasksKnown {
			pending = "+"
		}
		// fmt.Print("\033[u\033[K")
		// fmt.Printf("%d%% %s\n", int(perc), pc.Message)
		fmt.Printf("%d%% (%d/%d%s) %s\n", int(perc), pc.CompletedTasks, pc.TotalTasks, pending, pc.Message)
	}
}
