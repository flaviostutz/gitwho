package ownership

import (
	"fmt"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

// ServeOwnershipTimeseries Start server with a web page with graphs and
// returns the random URL generated for the page
func ServeOwnershipTimeseries(ownershipResults []OwnershipResult, ownershipTimeseriesOpts OwnershipTimeseriesOptions) string {

	// OWNERSHIP SHARE TIMESERIES
	tr := charts.NewThemeRiver()
	tr.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine}),
		charts.WithTitleOpts(opts.Title{
			Title: "Line Ownership Timeseries",
		}),
		charts.WithSingleAxisOpts(opts.SingleAxis{
			Type:   "time",
			Bottom: "10%",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
			Type: "scroll",
			Top:  "23px",
		}),
	)

	data := make([]opts.ThemeRiverData, 0)
	for _, resultsTs := range ownershipResults {
		for _, authorLines := range resultsTs.AuthorsLines {
			data = append(data, opts.ThemeRiverData{
				Date:  resultsTs.Commit.Date.Format(time.RFC3339),
				Name:  authorLines.AuthorName,
				Value: float64(authorLines.OwnedLinesTotal),
			})
		}
	}

	tr.AddSeries("ownership", data)

	// OWNERSHIP TOTAL PER AUTHOR TIMESERIES
	datesX := make([]string, 0)
	for _, oresults := range ownershipResults {
		datesX = append(datesX, oresults.Commit.Date.Format(time.DateOnly))
	}
	lineAuthor := charts.NewLine()
	lineAuthor.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine}),
		charts.WithTitleOpts(opts.Title{
			Title: "Line Ownership per Author",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "category",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
			Type: "scroll",
			Top:  "23px",
		}),
		charts.WithAnimation(),
	)

	lineAuthor.SetXAxis(datesX)
	byAuthorResults := SortByAuthorDate(ownershipResults)
	for _, authorResult := range byAuthorResults {
		authorValues := make([]opts.LineData, 0)
		for _, date := range datesX {
			authorValue := 0
			// look for value on this date
			for _, authorResult := range authorResult.AuthorLinesDate {
				if authorResult.Date == date {
					authorValue = authorResult.AuthorLines.OwnedLinesTotal
					break
				}
			}
			authorValues = append(authorValues, opts.LineData{Value: authorValue})
		}
		// for i, authorResultsDate := range authorResult.AuthorLinesDate {
		// 	if authorResultsDate.Date == datesX[i] {
		// 		authorValues = append(authorValues, opts.LineData{Value: authorResultsDate.AuthorLines.OwnedLinesTotal})
		// 	} else {
		// 		authorValues = append(authorValues, opts.LineData{Value: 100})
		// 	}
		// }
		lineAuthor.AddSeries(authorResult.AuthorName, authorValues,
			charts.WithLineChartOpts(
				opts.LineChart{Smooth: false},
			),
		)
	}

	// TOTAL LINES
	// OWNERSHIP TOTAL PER AUTHOR TIMESERIES
	lineTotal := charts.NewLine()
	lineTotal.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine, Height: "350px"}),
		charts.WithTitleOpts(opts.Title{
			Title: "Lines",
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Type: "category",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
			Type: "scroll",
			Top:  "23px",
		}),
		charts.WithAnimation(),
	)

	lineTotal.SetXAxis(datesX)
	totalValues := make([]opts.LineData, 0)
	duplicateValues := make([]opts.LineData, 0)
	for _, date := range datesX {
		totalValue := 0
		duplicateValue := 0
		// look for value on this date
		for _, ownershipResult := range ownershipResults {
			if ownershipResult.Commit.Date.Format(time.DateOnly) == date {
				totalValue = ownershipResult.TotalLines
				duplicateValue = ownershipResult.TotalLinesDuplicated
				break
			}
		}
		totalValues = append(totalValues, opts.LineData{Value: totalValue})
		duplicateValues = append(duplicateValues, opts.LineData{Value: duplicateValue})
	}
	lineTotal.AddSeries("Total Lines", totalValues,
		charts.WithLineChartOpts(
			opts.LineChart{Smooth: false},
		),
	)
	lineTotal.AddSeries("Duplicate Lines", duplicateValues,
		charts.WithLineChartOpts(
			opts.LineChart{Smooth: false},
		),
	)

	page := components.NewPage()
	page.SetLayout(components.PageFlexLayout)
	page.AddCharts(
		tr, lineTotal, lineAuthor,
	)

	info := "<pre style=\"display:flex;justify-content:center\"><code>"
	info += utils.BaseOptsStr(ownershipTimeseriesOpts.BaseOptions)
	info += ownershipTimeseriesOptsStr(ownershipTimeseriesOpts)
	info += FormatTimeseriesOwnershipResults(ownershipResults, true)
	info += "</code></pre>"

	url, _ := utils.ServeGraphPage(page, info)
	return url
}

func ServeOwnership(ownershipResult OwnershipResult, ownershipOpts OwnershipOptions) string {
	pie := charts.NewPie()

	items := make([]opts.PieData, 0)
	for _, authorLines := range ownershipResult.AuthorsLines {
		items = append(items, opts.PieData{Name: authorLines.AuthorName, Value: authorLines.OwnedLinesTotal})
	}

	pie.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine}),
		charts.WithTitleOpts(opts.Title{
			Title: "Ownership",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
			Show:    true,
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
			Type: "scroll",
			Top:  "23px",
		}),
	)

	pie.AddSeries("pie", items).
		SetSeriesOptions(charts.WithLabelOpts(
			opts.Label{
				Show:      true,
				Formatter: "{b}: {c}",
			}),
		)

	page := components.NewPage()
	page.SetLayout(components.PageFlexLayout)
	page.AddCharts(pie)

	info := "<pre style=\"display:flex;justify-content:center\"><code>"
	info += utils.BaseOptsStr(ownershipOpts.BaseOptions)
	info += ownershipOptsStr(ownershipOpts)
	info += FormatCodeOwnershipResults(ownershipResult, true)
	info += "</code></pre>"

	url, _ := utils.ServeGraphPage(page, info)
	return url
}

func ownershipTimeseriesOptsStr(opts OwnershipTimeseriesOptions) string {
	str := utils.AttrStr("since", opts.Since)
	str += utils.AttrStr("until", opts.Until)
	str += utils.AttrStr("period", opts.Period)
	str += utils.AttrStr("min-duplicate", fmt.Sprintf("%d", opts.MinDuplicateLines))
	return str
}

func ownershipOptsStr(opts OwnershipOptions) string {
	str := utils.AttrStr("commit-id", opts.CommitId)
	str += utils.AttrStr("min-duplicate", fmt.Sprintf("%d", opts.MinDuplicateLines))
	return str
}
