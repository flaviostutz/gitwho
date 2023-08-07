package changes

import (
	"fmt"
)

func FormatTextResults(changesResult ChangesResult, opts ChangesOptions) string {
	if changesResult.TotalCommits == 0 {
		return "No commits found"
	}

	text := fmt.Sprintf("Total authors: %d\n", len(changesResult.AuthorsLines))
	text += fmt.Sprintf("Total files: %d\n", changesResult.TotalFiles)
	text += fmt.Sprintf("Total lines: %v\n", changesResult.TotalLines)

	text += "\n"
	for _, authorLines := range changesResult.AuthorsLines {
		text += fmt.Sprintf("%v", authorLines)
	}
	return text
}
