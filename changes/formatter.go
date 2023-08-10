package changes

import (
	"fmt"
)

func FormatTextResults(changes ChangesResult, opts ChangesOptions) string {
	if changes.TotalCommits == 0 {
		return "No commits found"
	}

	text := fmt.Sprintf("Total authors active: %d\n", len(changes.AuthorsLines))
	text += fmt.Sprintf("Total files touched: %d\n", changes.TotalFiles)
	text += fmt.Sprintf("Total lines touched: %d (avg age: %d days)\n", changes.TotalLines.New+changes.TotalLines.ChurnOther+changes.TotalLines.ChurnOwn+changes.TotalLines.RefactorOther+changes.TotalLines.RefactorOwn, (int(changes.TotalLines.AgeSum.Hours())/changes.TotalLines.Changes)/24)
	text += FormatLinesChanges(changes.TotalLines, LinesChanges{})

	for _, authorLines := range changes.AuthorsLines {
		text += fmt.Sprintf("\nAuthor: %s\n", authorLines.AuthorName)
		text += FormatLinesChanges(authorLines.Lines, changes.TotalLines)
		text += fmt.Sprintf("     * Churn done by others to own lines (help received): %d%s\n", authorLines.Lines.ChurnReceived, calcPerc(authorLines.Lines.ChurnReceived, changes.TotalLines.ChurnReceived))
	}
	return text
}

func FormatLinesChanges(changes LinesChanges, totals LinesChanges) string {
	text := fmt.Sprintf(" - New lines: %d%s\n", changes.New, calcPerc(changes.New, totals.New))
	text += fmt.Sprintf(" - Changed lines: %d%s\n", changes.Changes, calcPerc(changes.Changes, totals.Changes))
	text += fmt.Sprintf("   - Refactor: %d%s\n", changes.RefactorOwn+changes.RefactorOther, calcPerc(changes.RefactorOwn+changes.RefactorOther, totals.RefactorOwn+totals.RefactorOther))
	text += fmt.Sprintf("     - Refactor of own lines: %d%s\n", changes.RefactorOwn, calcPerc(changes.RefactorOwn, totals.RefactorOwn))
	text += fmt.Sprintf("     - Refactor of other's lines: %d%s\n", changes.RefactorOther, calcPerc(changes.RefactorOther, totals.RefactorOther))
	text += fmt.Sprintf("   - Churn: %d%s\n", changes.ChurnOwn+changes.ChurnOther, calcPerc(changes.ChurnOwn+changes.ChurnOther, totals.ChurnOwn+totals.ChurnOther))
	text += fmt.Sprintf("     - Churn of own lines: %d%s\n", changes.ChurnOwn, calcPerc(changes.ChurnOwn, totals.ChurnOwn))
	text += fmt.Sprintf("     - Churn of other's lines (help given): %d%s\n", changes.ChurnOther, calcPerc(changes.ChurnOther, totals.ChurnOther))
	return text
}

func calcPerc(value int, total int) string {
	if value == 0 || total == 0 {
		return ""
	}
	return fmt.Sprintf(" (%d%%)", int(100*float64(value)/float64(total)))
}
