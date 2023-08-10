package changes

import (
	"fmt"
	"sort"
)

func FormatFullTextResults(changes ChangesResult, opts ChangesOptions) string {
	if changes.TotalCommits == 0 {
		return "No changes found"
	}

	avgAgeStr := ""
	if changes.TotalLines.Changes > 0 {
		avgAgeStr = fmt.Sprintf(" (avg age: %d days)", (int(changes.TotalLines.AgeSum.Hours())/changes.TotalLines.Changes)/24)
	}

	text := fmt.Sprintf("Total authors active: %d\n", len(changes.AuthorsLines))
	text += fmt.Sprintf("Total files touched: %d\n", changes.TotalFiles)
	text += fmt.Sprintf("Total lines touched: %d%s\n", changes.TotalLines.New+changes.TotalLines.ChurnOther+changes.TotalLines.ChurnOwn+changes.TotalLines.RefactorOther+changes.TotalLines.RefactorOwn, avgAgeStr)
	text += FormatLinesChanges(changes.TotalLines, LinesChanges{})

	for _, authorLines := range changes.AuthorsLines {
		text += fmt.Sprintf("\nAuthor: %s\n", authorLines.AuthorName)
		text += FormatLinesChanges(authorLines.Lines, changes.TotalLines)
		text += fmt.Sprintf("     * Churn done by others to own lines (help received): %d%s\n", authorLines.Lines.ChurnReceived, calcPercStr(authorLines.Lines.ChurnReceived, changes.TotalLines.ChurnReceived))
	}
	return text
}

func FormatTopTextResults(changes ChangesResult, opts ChangesOptions) string {
	if changes.TotalCommits == 0 {
		return "No changes found"
	}

	// top new liners
	sort.Slice(changes.AuthorsLines, func(i, j int) bool {
		ai := changes.AuthorsLines[i].Lines
		aj := changes.AuthorsLines[j].Lines
		return ai.New > aj.New
	})
	text := "Top New Liners\n"
	for i := 0; i < len(changes.AuthorsLines) && i < 2; i++ {
		al := changes.AuthorsLines[i]
		text += fmt.Sprintf("  %s: %d%s\n", al.AuthorName, al.Lines.New, calcPercStr(al.Lines.New, changes.TotalLines.New))
	}

	// top refactorers
	sort.Slice(changes.AuthorsLines, func(i, j int) bool {
		ai := changes.AuthorsLines[i].Lines
		aj := changes.AuthorsLines[j].Lines
		return ai.RefactorOther+ai.RefactorOwn > aj.RefactorOther+aj.RefactorOwn
	})
	text += "\nTop Refactorers\n"
	for i := 0; i < len(changes.AuthorsLines) && i < 2; i++ {
		al := changes.AuthorsLines[i]
		text += fmt.Sprintf("  %s: %d%s\n", al.AuthorName, al.Lines.RefactorOther+al.Lines.RefactorOwn, calcPercStr(al.Lines.RefactorOther+al.Lines.RefactorOwn, changes.TotalLines.RefactorOther+changes.TotalLines.RefactorOwn))
	}

	// top helpers
	sort.Slice(changes.AuthorsLines, func(i, j int) bool {
		ai := changes.AuthorsLines[i].Lines
		aj := changes.AuthorsLines[j].Lines
		return ai.ChurnOther > aj.ChurnOther
	})
	text += "\nTop Helpers\n"
	for i := 0; i < len(changes.AuthorsLines) && i < 2; i++ {
		al := changes.AuthorsLines[i]
		text += fmt.Sprintf("  %s: %d%s\n", al.AuthorName, al.Lines.ChurnOther, calcPercStr(al.Lines.ChurnOther, changes.TotalLines.ChurnOther))
	}

	// top churners
	sort.Slice(changes.AuthorsLines, func(i, j int) bool {
		ai := changes.AuthorsLines[i].Lines
		aj := changes.AuthorsLines[j].Lines
		return ai.ChurnReceived+ai.ChurnOwn > aj.ChurnReceived+aj.ChurnOwn
	})
	text += "\nTop Churners\n"
	for i := 0; i < len(changes.AuthorsLines) && i < 2; i++ {
		al := changes.AuthorsLines[i]
		text += fmt.Sprintf("  %s: %d%s\n", al.AuthorName, al.Lines.ChurnOwn+al.Lines.ChurnReceived, calcPercStr(al.Lines.ChurnOwn+al.Lines.ChurnReceived, changes.TotalLines.ChurnOwn+changes.TotalLines.ChurnReceived))
	}

	// top coders
	sort.Slice(changes.AuthorsLines, func(i, j int) bool {
		ai := changes.AuthorsLines[i].Lines
		aj := changes.AuthorsLines[j].Lines
		return calcTopCoderScore(ai) > calcTopCoderScore(aj)
	})
	text += "\nTop Coders (new + refactor - churn)\n"
	for i := 0; i < len(changes.AuthorsLines) && i < 2; i++ {
		al := changes.AuthorsLines[i]
		text += fmt.Sprintf("  %s: %d%s\n", al.AuthorName, calcTopCoderScore(al.Lines), calcPercStr(calcTopCoderScore(al.Lines), calcTopCoderScore(changes.TotalLines)))
	}

	return text
}

func calcTopCoderScore(ai LinesChanges) int {
	return ai.New + 3*ai.RefactorOther + 2*ai.RefactorOwn - ai.ChurnOwn - 3*ai.ChurnOther - 2*ai.ChurnReceived
}

func FormatLinesChanges(changes LinesChanges, totals LinesChanges) string {
	text := fmt.Sprintf(" - New lines: %d%s\n", changes.New, calcPercStr(changes.New, totals.New))
	text += fmt.Sprintf(" - Changed lines: %d%s\n", changes.Changes, calcPercStr(changes.Changes, totals.Changes))
	text += fmt.Sprintf("   - Refactor: %d%s\n", changes.RefactorOwn+changes.RefactorOther, calcPercStr(changes.RefactorOwn+changes.RefactorOther, totals.RefactorOwn+totals.RefactorOther))
	text += fmt.Sprintf("     - Refactor of own lines: %d%s\n", changes.RefactorOwn, calcPercStr(changes.RefactorOwn, totals.RefactorOwn))
	text += fmt.Sprintf("     - Refactor of other's lines: %d%s\n", changes.RefactorOther, calcPercStr(changes.RefactorOther, totals.RefactorOther))
	text += fmt.Sprintf("   - Churn: %d%s\n", changes.ChurnOwn+changes.ChurnOther, calcPercStr(changes.ChurnOwn+changes.ChurnOther, totals.ChurnOwn+totals.ChurnOther))
	text += fmt.Sprintf("     - Churn of own lines: %d%s\n", changes.ChurnOwn, calcPercStr(changes.ChurnOwn, totals.ChurnOwn))
	text += fmt.Sprintf("     - Churn of other's lines (help given): %d%s\n", changes.ChurnOther, calcPercStr(changes.ChurnOther, totals.ChurnOther))
	return text
}

func calcPercStr(value int, total int) string {
	if value == 0 || total == 0 {
		return ""
	}
	return fmt.Sprintf(" (%d%%)", int(100*float64(value)/float64(total)))
}
