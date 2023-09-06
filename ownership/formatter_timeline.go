package ownership

import (
	"bytes"
	"fmt"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/rodaine/table"
)

func FormatTimeseriesOwnershipResults(ownershipResults []OwnershipResult, full bool) string {
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
			fmt.Sprintf("%d%s", result.TotalLines, utils.CalcDiffStr(result.TotalLines, prevResult.TotalLines)),
			fmt.Sprintf("%d%s", result.TotalLinesDuplicated, utils.CalcDiffStr(result.TotalLinesDuplicated, prevResult.TotalLinesDuplicated)),
			fmt.Sprintf("%d%s", result.TotalFiles, utils.CalcDiffStr(result.TotalFiles, prevResult.TotalFiles)))
		prevResult = result
	}
	tbl.AddRow("Inc/period",
		fmt.Sprintf("%d%s", prevResult.TotalLines-firstResult.TotalLines, utils.CalcDiffPercStr(prevResult.TotalLines, firstResult.TotalLines)),
		fmt.Sprintf("%d%s", prevResult.TotalLinesDuplicated-firstResult.TotalLinesDuplicated, utils.CalcDiffPercStr(prevResult.TotalLinesDuplicated, firstResult.TotalLinesDuplicated)),
		fmt.Sprintf("%d%s", prevResult.TotalFiles-firstResult.TotalFiles, utils.CalcDiffPercStr(prevResult.TotalFiles, firstResult.TotalFiles)),
	)
	tbl.Print()
	str += tblWriter.String()

	if full {
		str += formatAuthorsTimeseries(ownershipResults)
	}

	return str
}

func formatAuthorsTimeseries(ownershipResults []OwnershipResult) string {

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
