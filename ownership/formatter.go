package ownership

import (
	"fmt"
	"strconv"
)

func FormatTextResults(ownershipResult OwnershipResult, opts OwnershipOptions, showEmail bool) string {
	text := fmt.Sprintf("Total authors: %d\n", len(ownershipResult.AuthorsLines))
	text += fmt.Sprintf("Total files: %d\n", ownershipResult.TotalFiles)
	text += fmt.Sprintf("Avg line age: %s\n", avgLineAgeStr(ownershipResult.LinesAgeDaysSum, ownershipResult.TotalLines))
	text += fmt.Sprintf("Total lines: %d\n", ownershipResult.TotalLines)
	for _, authorLines := range ownershipResult.AuthorsLines {
		mailStr := ""
		if showEmail {
			mailStr = fmt.Sprintf(" %s", authorLines.AuthorMail)
		}
		text += fmt.Sprintf("  %s%s: %d (%s%%) %s\n", authorLines.AuthorName, mailStr, authorLines.OwnedLines, strconv.FormatFloat(float64(100)*(float64(authorLines.OwnedLines)/float64(ownershipResult.TotalLines)), 'f', 1, 32), avgLineAgeStr(authorLines.OwnedLinesAgeDaysSum, authorLines.OwnedLines))
	}
	return text
}

func avgLineAgeStr(linesAgeDaysSum float64, totalLines int) string {
	// fmt.Printf("%s %d\n", linesAgeSum, totalLines)
	return fmt.Sprintf("%1.f days", (linesAgeDaysSum / float64(totalLines)))
}
