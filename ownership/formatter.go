package ownership

import (
	"fmt"
	"strconv"
	"strings"
)

func FormatTextResults(ownershipResult OwnershipResult, opts OwnershipOptions) string {
	text := fmt.Sprintf("Total authors: %d\n", len(ownershipResult.AuthorsLines))
	text += fmt.Sprintf("Total files: %d\n", ownershipResult.TotalFiles)
	text += fmt.Sprintf("Avg line age: %s\n", avgLineAgeStr(ownershipResult.LinesAgeDaysSum, ownershipResult.TotalLines))
	text += fmt.Sprintf("Duplicated lines: %d (%d%%)\n", ownershipResult.TotalLinesDuplicated, int(100*float64(ownershipResult.TotalLinesDuplicated)/float64(ownershipResult.TotalLines)))
	text += fmt.Sprintf("Total lines: %d\n", ownershipResult.TotalLines)
	for _, authorLines := range ownershipResult.AuthorsLines {
		if strings.Trim(authorLines.AuthorName, " ") == "" {
			fmt.Printf("BBB%sBBBB\n", authorLines.AuthorName)
		}
		mailStr := fmt.Sprintf(" %s", authorLines.AuthorMail)
		text += fmt.Sprintf("  %s%s: %d (%s%%) avg-days:%d dup:%d orig:%d dup-others:%d\n",
			authorLines.AuthorName, mailStr,
			authorLines.OwnedLinesTotal,
			strconv.FormatFloat(float64(100)*(float64(authorLines.OwnedLinesTotal)/float64(ownershipResult.TotalLines)), 'f', 1, 32),
			int((authorLines.OwnedLinesAgeDaysSum / float64(authorLines.OwnedLinesTotal))),
			authorLines.OwnedLinesDuplicate,
			authorLines.OwnedLinesDuplicateOriginal,
			authorLines.OwnedLinesDuplicateOriginalOthers)
	}
	return text
}

func avgLineAgeStr(linesAgeDaysSum float64, totalLines int) string {
	// fmt.Printf("%s %d\n", linesAgeSum, totalLines)
	return fmt.Sprintf("%1.f days", (linesAgeDaysSum / float64(totalLines)))
}
