package ownership

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/rodaine/table"
)

type authorLinesDate struct {
	date        string
	authorLines AuthorLines
}

func FormatTimelineOwnershipResults(ownershipResults []OwnershipResult, full bool) string {
	str := "\n"

	tblWriter := bytes.NewBufferString("")
	tbl := table.New("Date", "Lines", "Duplicates", "Files")
	tbl.WithWriter(tblWriter)

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
	str += tblWriter.String()

	if full {
		str += formatAuthorsTimelines(ownershipResults)
	}

	return str
}

func formatAuthorsTimelines(ownershipResults []OwnershipResult) string {

	str := ""
	authorNameLinesDates := SortByAuthorDate(ownershipResults)

	// display data
	for _, authorNameLinesDate := range authorNameLinesDates {
		str += fmt.Sprintf("\n%s\n", authorNameLinesDate.AuthorName)

		tblWriter := bytes.NewBufferString("")
		tbl := table.New("Date", "Lines", "Duplicates (total)", "Duplicates (original)")
		tbl.WithWriter(tblWriter)

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
		str += tblWriter.String()
	}
	return str
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
