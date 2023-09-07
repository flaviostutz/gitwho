package changes

import (
	"bytes"
	"fmt"
	"time"

	"github.com/flaviostutz/gitwho/changes"
	"github.com/flaviostutz/gitwho/utils"
	"github.com/rodaine/table"
)

func FormatTimeseriesChangesResults(changesResults []changes.ChangesResult, full bool) (string, error) {
	str := "\n"

	tblWriter := bytes.NewBufferString("")
	tbl := table.New("Period", "Commits", "Files touched", "Lines touched", "New lines", "Changed lines")
	tbl.WithWriter(tblWriter)

	prevResult := changes.ChangesResult{}

	totalCommits := 0
	totalFiles := 0
	totalChanges := 0
	totalNew := 0

	for _, result := range changesResults {
		tbl.AddRow(fmt.Sprintf("%s - %s", result.SinceCommit.Date.Format(time.DateOnly), result.UntilCommit.Date.Format(time.DateOnly)),
			fmt.Sprintf("%d%s", result.TotalCommits, utils.CalcDiffStr(result.TotalCommits, prevResult.TotalCommits)),
			fmt.Sprintf("%d%s", result.TotalFiles, utils.CalcDiffStr(result.TotalFiles, prevResult.TotalFiles)),
			fmt.Sprintf("%d%s", totalTouched(result.TotalLinesTouched), utils.CalcDiffStr(totalTouched(result.TotalLinesTouched), totalTouched(prevResult.TotalLinesTouched))),
			fmt.Sprintf("%d%s", result.TotalLinesTouched.New, utils.CalcDiffStr(result.TotalLinesTouched.New, prevResult.TotalLinesTouched.New)),
			fmt.Sprintf("%d%s", result.TotalLinesTouched.Changes, utils.CalcDiffStr(result.TotalLinesTouched.Changes, prevResult.TotalLinesTouched.Changes)),
		)
		totalCommits += result.TotalCommits
		totalFiles += result.TotalFiles
		totalNew += result.TotalLinesTouched.New
		totalChanges += result.TotalLinesTouched.Changes
		prevResult = result
	}
	tbl.AddRow("Total",
		fmt.Sprintf("%d", totalCommits),
		fmt.Sprintf("%d", totalFiles),
		fmt.Sprintf("%d", totalNew+totalChanges),
		fmt.Sprintf("%d", totalNew),
		fmt.Sprintf("%d", totalChanges),
	)
	tbl.Print()
	str += tblWriter.String()

	// author clusters
	aclusters, err := changes.ClusterizeAuthors(changesResults, 3)
	if err != nil {
		return "", err
	}

	text := ""
	if len(aclusters) > 0 {
		text += "\nAuthor clusters\n"
		for _, authorCluster := range aclusters {
			authorNames := make([]string, 0)
			for _, lines := range authorCluster.Lines {
				authorNames = append(authorNames, lines.AuthorName)
			}
			text += fmt.Sprintf("  %s: %s\n", authorCluster.Name, utils.JoinWithLimit(authorNames, ", ", 4))
		}
	}

	if full {
		str += formatAuthorsTimeseries(changesResults)
	}

	return str, nil
}

func formatAuthorsTimeseries(changesResults []changes.ChangesResult) string {

	str := ""
	authorNameLinesDates := changes.SortByAuthorDate(changesResults)

	// display data
	for _, authorNameLinesDate := range authorNameLinesDates {
		str += fmt.Sprintf("\n%s\n", authorNameLinesDate.AuthorName)

		tblWriter := bytes.NewBufferString("")
		tbl := table.New("Period", "Lines touched", "New lines", "Changed lines")
		tbl.WithWriter(tblWriter)

		prevResult := changes.AuthorLinesDate{}

		totalChanges := 0
		totalNew := 0

		for _, result := range authorNameLinesDate.AuthorLinesDates {
			tbl.AddRow(fmt.Sprintf("%s", result.Since),
				fmt.Sprintf("%d%s", totalTouched(result.AuthorLines.LinesTouched), utils.CalcDiffStr(totalTouched(result.AuthorLines.LinesTouched), totalTouched(prevResult.AuthorLines.LinesTouched))),
				fmt.Sprintf("%d%s", result.AuthorLines.LinesTouched.New, utils.CalcDiffStr(result.AuthorLines.LinesTouched.New, prevResult.AuthorLines.LinesTouched.New)),
				fmt.Sprintf("%d%s", result.AuthorLines.LinesTouched.Changes, utils.CalcDiffStr(result.AuthorLines.LinesTouched.Changes, prevResult.AuthorLines.LinesTouched.Changes)),
			)
			totalNew += result.AuthorLines.LinesTouched.New
			totalChanges += result.AuthorLines.LinesTouched.Changes
			prevResult = result
		}
		tbl.AddRow("Total",
			fmt.Sprintf("%d", totalNew+totalChanges),
			fmt.Sprintf("%d", totalNew),
			fmt.Sprintf("%d", totalChanges),
		)
		tbl.Print()
		str += tblWriter.String()

	}
	return str
}

func totalTouched(linesTouched changes.LinesTouched) int {
	return linesTouched.New + linesTouched.Changes
}
