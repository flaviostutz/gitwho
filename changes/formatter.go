package changes

import (
	"fmt"
	"sort"
)

func FormatFullTextResults(changes ChangesResult, opts ChangesOptions) string {
	if changes.TotalCommits == 0 {
		return "No changes found"
	}

	text := fmt.Sprintf("Total authors active: %d\n", len(changes.AuthorsLines))
	text += fmt.Sprintf("Total files touched: %d\n", changes.TotalFiles)
	if changes.TotalLines.Changes > 0 {
		text += fmt.Sprintf("Average line age when changed: %d days\n", (int(changes.TotalLines.AgeDaysSum / float64(changes.TotalLines.Changes))))
	}
	text += formatLinesTouched(changes.TotalLines, LinesTouched{})

	for _, authorLines := range changes.AuthorsLines {
		if authorLines.LinesTouched.New+authorLines.LinesTouched.Changes == 0 {
			continue
		}
		mailStr := fmt.Sprintf(" %s", authorLines.AuthorMail)
		text += fmt.Sprintf("\nAuthor: %s%s\n", authorLines.AuthorName, mailStr)
		text += formatLinesTouched(authorLines.LinesTouched, changes.TotalLines)
		text += formatTopTouchedFiles(authorLines.FilesTouched)
	}
	return text
}

func FormatTopTextResults(changes ChangesResult, opts ChangesOptions) string {
	if changes.TotalCommits == 0 {
		return "No changes found"
	}

	// top coders
	sort.Slice(changes.AuthorsLines, func(i, j int) bool {
		ai := changes.AuthorsLines[i].LinesTouched
		aj := changes.AuthorsLines[j].LinesTouched
		return calcTopCoderScore(ai) > calcTopCoderScore(aj)
	})
	text := "Top Coders (new+refactor-churn)\n"
	for i := 0; i < len(changes.AuthorsLines) && i < 3; i++ {
		al := changes.AuthorsLines[i]
		mailStr := fmt.Sprintf(" %s", al.AuthorMail)
		text += fmt.Sprintf("  %s%s: %d%s\n", al.AuthorName, mailStr, calcTopCoderScore(al.LinesTouched), calcPercStr(calcTopCoderScore(al.LinesTouched), calcTopCoderScore(changes.TotalLines)))
	}

	// top new liners
	sort.Slice(changes.AuthorsLines, func(i, j int) bool {
		ai := changes.AuthorsLines[i].LinesTouched
		aj := changes.AuthorsLines[j].LinesTouched
		return ai.New > aj.New
	})
	text += "\nTop New Liners\n"
	for i := 0; i < len(changes.AuthorsLines) && i < 3; i++ {
		al := changes.AuthorsLines[i]
		text += fmt.Sprintf("  %s: %d%s\n", al.AuthorName, al.LinesTouched.New, calcPercStr(al.LinesTouched.New, changes.TotalLines.New))
	}

	// top refactorers
	sort.Slice(changes.AuthorsLines, func(i, j int) bool {
		ai := changes.AuthorsLines[i].LinesTouched
		aj := changes.AuthorsLines[j].LinesTouched
		return ai.RefactorOther+ai.RefactorOwn > aj.RefactorOther+aj.RefactorOwn
	})
	text += "\nTop Refactorers\n"
	for i := 0; i < len(changes.AuthorsLines) && i < 3; i++ {
		al := changes.AuthorsLines[i]
		text += fmt.Sprintf("  %s: %d%s\n", al.AuthorName, al.LinesTouched.RefactorOther+al.LinesTouched.RefactorOwn, calcPercStr(al.LinesTouched.RefactorOther+al.LinesTouched.RefactorOwn, changes.TotalLines.RefactorOther+changes.TotalLines.RefactorOwn))
	}

	// top helpers
	sort.Slice(changes.AuthorsLines, func(i, j int) bool {
		ai := changes.AuthorsLines[i].LinesTouched
		aj := changes.AuthorsLines[j].LinesTouched
		return ai.ChurnOther > aj.ChurnOther
	})
	text += "\nTop Helpers\n"
	for i := 0; i < len(changes.AuthorsLines) && i < 3; i++ {
		al := changes.AuthorsLines[i]
		text += fmt.Sprintf("  %s: %d%s\n", al.AuthorName, al.LinesTouched.ChurnOther, calcPercStr(al.LinesTouched.ChurnOther, changes.TotalLines.ChurnOther))
	}

	// top churners
	sort.Slice(changes.AuthorsLines, func(i, j int) bool {
		ai := changes.AuthorsLines[i].LinesTouched
		aj := changes.AuthorsLines[j].LinesTouched
		return ai.ChurnReceived+ai.ChurnOwn > aj.ChurnReceived+aj.ChurnOwn
	})
	text += "\nTop Churners\n"
	for i := 0; i < len(changes.AuthorsLines) && i < 3; i++ {
		al := changes.AuthorsLines[i]
		text += fmt.Sprintf("  %s: %d%s\n", al.AuthorName, al.LinesTouched.ChurnOwn+al.LinesTouched.ChurnReceived, calcPercStr(al.LinesTouched.ChurnOwn+al.LinesTouched.ChurnReceived, changes.TotalLines.ChurnOwn+changes.TotalLines.ChurnReceived))
	}

	return text
}

func formatTopTouchedFiles(filesTouched []FileTouched) string {
	text := fmt.Sprintf("  - Top files:\n")
	sort.Slice(filesTouched, func(i, j int) bool {
		return filesTouched[i].Lines > filesTouched[j].Lines
	})
	for i := 0; i < len(filesTouched) && i < 5; i++ {
		text += fmt.Sprintf("    - %s (%d)\n", filesTouched[i].Name, filesTouched[i].Lines)
	}
	return text
}

func calcTopCoderScore(ai LinesTouched) int {
	return ai.New + 3*ai.RefactorOther + 2*ai.RefactorOwn - 2*ai.ChurnOwn - 4*ai.ChurnReceived
}

func formatLinesTouched(changes LinesTouched, totals LinesTouched) string {
	totalTouched := changes.New + changes.Changes
	text := fmt.Sprintf("- Total lines touched: %d%s\n", totalTouched, calcPercStr(changes.New+changes.Changes, totals.New+totals.Changes))
	text += fmt.Sprintf("  - New lines: %d%s\n", changes.New, calcPercStr(changes.New, totalTouched))
	text += fmt.Sprintf("  - Changed lines: %d%s\n", changes.Changes, calcPercStr(changes.Changes, totalTouched))
	text += fmt.Sprintf("    - Refactor: %d%s\n", changes.RefactorOwn+changes.RefactorOther, calcPercStr(changes.RefactorOwn+changes.RefactorOther, changes.Changes))
	text += fmt.Sprintf("      - Refactor of own lines: %d%s\n", changes.RefactorOwn, calcPercStr(changes.RefactorOwn, changes.RefactorOwn+changes.RefactorOther))
	text += fmt.Sprintf("      - Refactor of other's lines: %d%s\n", changes.RefactorOther, calcPercStr(changes.RefactorOther, changes.RefactorOwn+changes.RefactorOther))
	text += fmt.Sprintf("      * Refactor done by others to own lines (help received): %d\n", changes.RefactorReceived)
	text += fmt.Sprintf("    - Churn: %d%s\n", changes.ChurnOwn+changes.ChurnOther, calcPercStr(changes.ChurnOwn+changes.ChurnOther, changes.Changes))
	text += fmt.Sprintf("      - Churn of own lines: %d%s\n", changes.ChurnOwn, calcPercStr(changes.ChurnOwn, changes.ChurnOwn+changes.ChurnOther))
	text += fmt.Sprintf("      - Churn of other's lines (help given): %d%s\n", changes.ChurnOther, calcPercStr(changes.ChurnOther, changes.ChurnOwn+changes.ChurnOther))
	text += fmt.Sprintf("      * Churn done by others to own lines (help received): %d\n", changes.ChurnReceived)
	return text
}

func calcPercStr(value int, total int) string {
	if total == 0 {
		return ""
	}
	return fmt.Sprintf(" (%d%%)", int(100*float64(value)/float64(total)))
}
