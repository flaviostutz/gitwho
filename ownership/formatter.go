package ownership

import (
	"fmt"
	"strconv"
)

func FormatTextResults(ownershipResult OwnershipResult, opts OwnershipOptions) string {
	text := fmt.Sprintf("Total authors: %d\n", len(ownershipResult.AuthorsLines))
	text += fmt.Sprintf("Total files: %d\n", ownershipResult.TotalFiles)
	text += fmt.Sprintf("Total lines: %d\n", ownershipResult.TotalLines)
	text += "\n"
	for _, authorLines := range ownershipResult.AuthorsLines {
		text += fmt.Sprintf("%s: %d (%s%%)\n", authorLines.Author, authorLines.OwnedLines, strconv.FormatFloat(float64(100)*(float64(authorLines.OwnedLines)/float64(ownershipResult.TotalLines)), 'f', 1, 32))
	}
	return text
}
