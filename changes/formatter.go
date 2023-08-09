package changes

import (
	"fmt"
)

func FormatTextResults(changes ChangesResult, opts ChangesOptions) string {
	if changes.TotalCommits == 0 {
		return "No commits found"
	}

	text := fmt.Sprintf("Total authors active: %d\n", len(changes.AuthorsLines))
	text += fmt.Sprintf("Total files changed: %d\n", changes.TotalFiles)
	text += fmt.Sprintf("Total lines changed: %d (avg age: %d days)\n", changes.TotalLines.New+changes.TotalLines.ChurnOther+changes.TotalLines.ChurnOwn+changes.TotalLines.RefactorOther+changes.TotalLines.RefactorOwn, (int(changes.TotalLines.AgeSum.Hours())/changes.TotalLines.Changes)/24)
	text += FormatLinesChanges(changes.TotalLines)

	text += "\n"
	for _, authorLines := range changes.AuthorsLines {
		text += fmt.Sprintf("Author: %s\n", authorLines.AuthorName)
		text += FormatLinesChanges(authorLines.Lines)
		text += fmt.Sprintf("    Help received: %d\n", authorLines.Lines.ChurnReceived)
	}
	return text
}

func FormatLinesChanges(changes LinesChanges) string {
	text := fmt.Sprintf(" - New lines: %d\n", changes.New)
	text += fmt.Sprintf(" - Changed lines: %d\n", changes.Changes)
	text += fmt.Sprintf("   - Refactor: %d\n", changes.RefactorOwn+changes.RefactorOther)
	text += fmt.Sprintf("     - Own lines: %d\n", changes.RefactorOwn)
	text += fmt.Sprintf("     - Other's lines: %d\n", changes.RefactorOther)
	text += fmt.Sprintf("   - Churn: %d\n", changes.ChurnOwn)
	text += fmt.Sprintf("     - Own lines: %d\n", changes.ChurnOwn)
	text += fmt.Sprintf("     - Other's lines: %d\n", changes.ChurnOther)
	return text
}
