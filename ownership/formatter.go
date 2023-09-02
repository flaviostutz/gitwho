package ownership

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/flaviostutz/gitwho/utils"
	"github.com/rodaine/table"
)

type authorLinesDate struct {
	date        string
	authorLines AuthorLines
}

func PrintTimelineOwnershipResults(ownershipResults []OwnershipResult, full bool) {
	fmt.Println()

	headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
	columnFmt := color.New(color.FgYellow).SprintfFunc()

	tbl := table.New("Date", "Lines", "Duplicates", "Files")
	tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

	firstResult := OwnershipResult{}
	prevResult := OwnershipResult{}
	for i, result := range ownershipResults {
		if i == 0 {
			firstResult = result
		}
		tbl.AddRow(result.Commit.Date.Format(time.DateOnly),
			fmt.Sprintf("%d%s", result.TotalLinesDuplicated, utils.CalcDiffStr(result.TotalLinesDuplicated, prevResult.TotalLinesDuplicated)),
			fmt.Sprintf("%d%s", result.TotalLines, utils.CalcDiffStr(result.TotalLines, prevResult.TotalLines)),
			fmt.Sprintf("%d%s", result.TotalFiles, utils.CalcDiffStr(result.TotalFiles, prevResult.TotalFiles)))
		prevResult = result
	}
	tbl.AddRow("Inc/period",
		fmt.Sprintf("%d%s", prevResult.TotalLinesDuplicated-firstResult.TotalLinesDuplicated, utils.CalcDiffPercStr(prevResult.TotalLinesDuplicated, firstResult.TotalLinesDuplicated)),
		fmt.Sprintf("%d%s", prevResult.TotalLines-firstResult.TotalLines, utils.CalcDiffPercStr(prevResult.TotalLines, firstResult.TotalLines)),
		fmt.Sprintf("%d%s", prevResult.TotalFiles-firstResult.TotalFiles, utils.CalcDiffPercStr(prevResult.TotalFiles, firstResult.TotalFiles)),
	)
	tbl.Print()

	if full {
		printAuthorsTimelines(ownershipResults)
	}
}

func printAuthorsTimelines(ownershipResults []OwnershipResult) {

	authorNameLinesDates := SortByAuthorDate(ownershipResults)

	// display data
	for _, authorNameLinesDate := range authorNameLinesDates {
		fmt.Printf("\n%s\n", authorNameLinesDate.AuthorName)
		headerFmt := color.New(color.FgGreen, color.Underline).SprintfFunc()
		columnFmt := color.New(color.FgYellow).SprintfFunc()

		tbl := table.New("Date", "Lines", "Duplicates (total)", "Duplicates (original)")
		tbl.WithHeaderFormatter(headerFmt).WithFirstColumnFormatter(columnFmt)

		firstResult := AuthorLines{}
		prevResult := AuthorLines{}

		for _, linesData := range authorNameLinesDate.AuthorLinesDate {
			result := linesData.AuthorLines
			if firstResult.AuthorName == "" {
				firstResult = result
			}
			tbl.AddRow(linesData.Date,
				fmt.Sprintf("%d%s", result.OwnedLinesTotal, utils.CalcDiffStr(result.OwnedLinesTotal, prevResult.OwnedLinesTotal)),
				fmt.Sprintf("%d%s", result.OwnedLinesDuplicate, utils.CalcDiffStr(result.OwnedLinesDuplicate, prevResult.OwnedLinesDuplicate)),
				fmt.Sprintf("%d%s", result.OwnedLinesDuplicateOriginal, utils.CalcDiffStr(result.OwnedLinesDuplicateOriginal, prevResult.OwnedLinesDuplicateOriginal)))
			prevResult = result
		}
		tbl.AddRow("Inc/period",
			fmt.Sprintf("%d%s", prevResult.OwnedLinesTotal-firstResult.OwnedLinesTotal, utils.CalcDiffPercStr(prevResult.OwnedLinesTotal, firstResult.OwnedLinesTotal)),
			fmt.Sprintf("%d%s", prevResult.OwnedLinesDuplicate-firstResult.OwnedLinesDuplicate, utils.CalcDiffPercStr(prevResult.OwnedLinesDuplicate, firstResult.OwnedLinesDuplicate)),
			fmt.Sprintf("%d%s", prevResult.OwnedLinesDuplicateOriginal-firstResult.OwnedLinesDuplicateOriginal, utils.CalcDiffPercStr(prevResult.OwnedLinesDuplicateOriginal, firstResult.OwnedLinesDuplicateOriginal)),
		)
		tbl.Print()
	}
}

func FormatCodeOwnershipResults(ownershipResult OwnershipResult, full bool) string {
	text := fmt.Sprintf("\nTotal authors: %d\n", len(ownershipResult.AuthorsLines))
	text += fmt.Sprintf("Total files: %d\n", ownershipResult.TotalFiles)
	if full {
		text += fmt.Sprintf("Avg line age: %s\n", avgLineAgeStr(ownershipResult.LinesAgeDaysSum, ownershipResult.TotalLines))
		text += fmt.Sprintf("Duplicated lines: %d (%d%%)\n", ownershipResult.TotalLinesDuplicated, int(100*float64(ownershipResult.TotalLinesDuplicated)/float64(ownershipResult.TotalLines)))
	}
	text += fmt.Sprintf("Total lines: %d\n", ownershipResult.TotalLines)
	for _, authorLines := range ownershipResult.AuthorsLines {
		mailStr := ""
		additional := ""
		if full {
			additional = fmt.Sprintf(" avg-days:%d dup:%d orig:%d dup-others:%d",
				int((authorLines.OwnedLinesAgeDaysSum / float64(authorLines.OwnedLinesTotal))),
				authorLines.OwnedLinesDuplicate,
				authorLines.OwnedLinesDuplicateOriginal,
				authorLines.OwnedLinesDuplicateOriginalOthers)
			mailStr = fmt.Sprintf(" %s", authorLines.AuthorMail)
		}
		text += fmt.Sprintf("  %s%s: %d (%s%%)%s\n",
			authorLines.AuthorName,
			mailStr,
			authorLines.OwnedLinesTotal,
			strconv.FormatFloat(float64(100)*(float64(authorLines.OwnedLinesTotal)/float64(ownershipResult.TotalLines)), 'f', 1, 32),
			additional)
	}

	return text
}

func FormatDuplicatesResults(ownershipResult OwnershipResult, full bool) string {
	text := fmt.Sprintf("Total lines: %d\n", ownershipResult.TotalLines)
	text += fmt.Sprintf("Duplicated lines: %d (%d%%)\n", ownershipResult.TotalLinesDuplicated, int(100*float64(ownershipResult.TotalLinesDuplicated)/float64(ownershipResult.TotalLines)))
	counter := 0
	for _, lineGroup := range ownershipResult.DuplicateLineGroups {
		text += fmt.Sprintf("%s:%d - %d\n", lineGroup.FilePath, lineGroup.LineNumber, lineGroup.LineNumber+lineGroup.LineCount)
		for _, relatedGroup := range lineGroup.RelatedLinesGroup {
			text += fmt.Sprintf("  %s:%d - %d\n", relatedGroup.FilePath, relatedGroup.LineNumber, relatedGroup.LineNumber+relatedGroup.LineCount)
			if !full {
				counter++
				if counter > 20 {
					text += "...(use --format \"full\" for more results)\n"
					return text
				}
			}
		}
	}
	return text
}

func avgLineAgeStr(linesAgeDaysSum float64, totalLines int) string {
	// fmt.Printf("%s %d\n", linesAgeSum, totalLines)
	return fmt.Sprintf("%1.f days", (linesAgeDaysSum / float64(totalLines)))
}

func formatSimpleOwnershipTimelineResults(results []OwnershipResult) string {
	str := "\nDate | Lines | Duplicates | Files\n"

	firstResult := OwnershipResult{}
	prevResult := OwnershipResult{}
	for i, result := range results {
		if i == 0 {
			firstResult = result
		}
		str += fmt.Sprintf("%s | %d%s | %d%s | %d%s\n",
			result.Commit.Date.Format(time.DateOnly),
			result.TotalLinesDuplicated, utils.CalcDiffStr(result.TotalLinesDuplicated, prevResult.TotalLinesDuplicated),
			result.TotalLines, utils.CalcDiffStr(result.TotalLines, prevResult.TotalLines),
			result.TotalFiles, utils.CalcDiffStr(result.TotalFiles, prevResult.TotalFiles),
		)
		prevResult = result
	}
	str += fmt.Sprintf("%s | %d%s | %d%s | %d%s\n",
		"Inc/period",
		prevResult.TotalLinesDuplicated-firstResult.TotalLinesDuplicated, utils.CalcDiffPercStr(prevResult.TotalLinesDuplicated, firstResult.TotalLinesDuplicated),
		prevResult.TotalLines-firstResult.TotalLines, utils.CalcDiffPercStr(prevResult.TotalLines, firstResult.TotalLines),
		prevResult.TotalFiles-firstResult.TotalFiles, utils.CalcDiffPercStr(prevResult.TotalFiles, firstResult.TotalFiles),
	)
	return str
}
