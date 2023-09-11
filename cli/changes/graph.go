package changes

import (
	"fmt"
	"time"

	"github.com/flaviostutz/gitwho/changes"
	"github.com/flaviostutz/gitwho/cli"
	"github.com/flaviostutz/gitwho/utils"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

// ServeChangesTimeseries Start server with a web page with graphs and
// returns the random URL generated for the page
func ServeChangesTimeseries(changesResults []changes.ChangesResult, ownershipTimeseriesOpts changes.ChangesTimeseriesOptions) string {

	// CHANGES TIMESERIES
	tr := charts.NewThemeRiver()
	tr.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme: types.ThemeShine,
		}),
		charts.WithTitleOpts(opts.Title{
			Title: "Lines Touched Timeseries",
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
	for _, resultsTs := range changesResults {
		for _, authorLines := range resultsTs.AuthorsLines {
			if authorLines.LinesTouched.New+authorLines.LinesTouched.Changes == 0 {
				continue
			}
			data = append(data, opts.ThemeRiverData{
				Date:  resultsTs.UntilCommit.Date.Format(time.RFC3339),
				Name:  authorLines.AuthorName,
				Value: float64(authorLines.LinesTouched.New + authorLines.LinesTouched.Changes),
			})
		}
	}

	tr.AddSeries("changes", data)

	// // TOTAL CHANGES PIE
	// pie := charts.NewPie()

	// authorTotals := make(map[string]int, 0)
	// items := make([]opts.PieData, 0)
	// for _, resultsTs := range changesResults {
	// 	for _, authorLines := range resultsTs.AuthorsLines {
	// 		if authorLines.LinesTouched.New+authorLines.LinesTouched.Changes == 0 {
	// 			continue
	// 		}
	// 		authorTotal := authorTotals[authorLines.AuthorName]
	// 		authorTotal += authorLines.LinesTouched.New + authorLines.LinesTouched.Changes
	// 		authorTotals[authorLines.AuthorName] = authorTotal
	// 	}
	// }

	// for authorName, authorTotal := range authorTotals {
	// 	items = append(items, opts.PieData{Name: authorName, Value: authorTotal})
	// }

	// pie.SetGlobalOptions(
	// 	charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine}),
	// 	charts.WithTitleOpts(opts.Title{
	// 		Title: "Total Lines Touched",
	// 	}),
	// 	charts.WithTooltipOpts(opts.Tooltip{
	// 		Trigger: "axis",
	// 		Show:    true,
	// 	}),
	// 	charts.WithLegendOpts(opts.Legend{
	// 		Show: true,
	// 		Type: "scroll",
	// 		Top:  "23px",
	// 	}),
	// )

	// pie.AddSeries("pie", items).
	// 	SetSeriesOptions(charts.WithLabelOpts(
	// 		opts.Label{
	// 			Show:      true,
	// 			Formatter: "{b}: {c}",
	// 		}),
	// 	)

	datesX := make([]string, 0)
	for _, oresults := range changesResults {
		datesX = append(datesX, oresults.UntilCommit.Date.Format(time.DateOnly))
	}

	// TOTAL LINES TOUCHED
	linesTouched := charts.NewLine()
	linesTouched.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  types.ThemeShine,
			Height: "250px"},
		),
		charts.WithTitleOpts(opts.Title{
			Title: "Lines Touched",
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

	linesTouched.SetXAxis(datesX)

	totalValues := make([]opts.LineData, 0)
	for _, date := range datesX {
		totalValue := 0
		// look for value on this date
		for _, changesResult := range changesResults {
			if changesResult.UntilCommit.Date.Format(time.DateOnly) == date {
				totalValue = changesResult.TotalLinesTouched.New + changesResult.TotalLinesTouched.Changes
				break
			}
		}
		totalValues = append(totalValues, opts.LineData{Value: totalValue})
	}
	linesTouched.AddSeries("Lines Touched", totalValues,
		charts.WithLineChartOpts(
			opts.LineChart{Smooth: false},
		),
	)

	// LINES TOUCHED PER AUTHOR TIMESERIES
	lineAuthor := charts.NewLine()
	lineAuthor.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Theme:  types.ThemeShine,
			Height: "250px"},
		),
		charts.WithTitleOpts(opts.Title{
			Title: "Lines Touched per Author",
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
	byAuthorResults := changes.SortByAuthorDate(changesResults)
	for _, authorResult := range byAuthorResults {
		authorValues := make([]opts.LineData, 0)
		for _, date := range datesX {
			authorValue := 0
			// look for value on this date
			for _, authorResult := range authorResult.AuthorLinesDates {
				if authorResult.Since == date {
					authorValue = authorResult.AuthorLines.LinesTouched.New + authorResult.AuthorLines.LinesTouched.Changes
					break
				}
			}
			authorValues = append(authorValues, opts.LineData{Value: authorValue})
		}
		lineAuthor.AddSeries(authorResult.AuthorName, authorValues,
			charts.WithLineChartOpts(
				opts.LineChart{Smooth: false},
			),
		)
	}

	// ADD GRAPHS TO PAGE
	page := components.NewPage()
	page.SetLayout(components.PageFlexLayout)
	page.AddCharts(tr, linesTouched, lineAuthor)

	info := "<pre style=\"display:flex;justify-content:center\"><code>"
	info += utils.BaseOptsStr(ownershipTimeseriesOpts.BaseOptions)
	info += changesTimeseriesOptsStr(ownershipTimeseriesOpts)

	tsresults, err := FormatTimeseriesChangesResults(changesResults, true)
	if err != nil {
		fmt.Printf("Couldn't format results. err=%s", err)
		panic(5)
	}
	info += tsresults
	info += "</code></pre>"

	url, _ := cli.ServeGraphPage(page, info)
	return url
}

// ServeChanges Start server with a web page with graphs and
// returns the random URL generated for the page
func ServeChanges(cresult changes.ChangesResult, changesOpts changes.ChangesOptions) (string, error) {

	sankey := charts.NewSankey()
	sankey.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine}),
		charts.WithTitleOpts(opts.Title{
			Title: "Changes",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: false,
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "item",
			Show:    true,
		}),
	)

	nodes := []opts.SankeyNode{
		{Name: "Lines touched"},
		{Name: "New lines"},
		{Name: "Changed lines"},
		{Name: "Refactor"},
		{Name: "Refactor own"},
		{Name: "Refactor others"},
		{Name: "Churn"},
		{Name: "Churn own"},
		{Name: "Churn others"},
	}

	links := make([]opts.SankeyLink, 0)
	links = append(links, opts.SankeyLink{Source: "Lines touched", Target: "Changed lines", Value: float32(cresult.TotalLinesTouched.Changes)})
	links = append(links, opts.SankeyLink{Source: "Changed lines", Target: "Refactor", Value: float32(cresult.TotalLinesTouched.RefactorOwn + cresult.TotalLinesTouched.RefactorOther)})
	links = append(links, opts.SankeyLink{Source: "Refactor", Target: "Refactor own", Value: float32(cresult.TotalLinesTouched.RefactorOwn)})
	links = append(links, opts.SankeyLink{Source: "Refactor", Target: "Refactor others", Value: float32(cresult.TotalLinesTouched.RefactorOther)})
	links = append(links, opts.SankeyLink{Source: "Changed lines", Target: "Churn", Value: float32(cresult.TotalLinesTouched.ChurnOwn + cresult.TotalLinesTouched.ChurnOther)})
	links = append(links, opts.SankeyLink{Source: "Churn", Target: "Churn own", Value: float32(cresult.TotalLinesTouched.ChurnOwn)})
	links = append(links, opts.SankeyLink{Source: "Churn", Target: "Churn others", Value: float32(cresult.TotalLinesTouched.ChurnOther)})
	links = append(links, opts.SankeyLink{Source: "Lines touched", Target: "New lines", Value: float32(cresult.TotalLinesTouched.New)})

	sankey.AddSeries("lines", nodes, links,
		charts.WithLabelOpts(opts.Label{
			Show: true,
		}),
	)

	page := components.NewPage()
	page.AddCharts(sankey)

	info := "<pre style=\"display:flex;justify-content:center\"><code>"
	info += utils.BaseOptsStr(changesOpts.BaseOptions)
	info += changesOptsStr(changesOpts)

	co, err := FormatFullTextResults(cresult)
	if err != nil {
		return "", err
	}
	info += co
	info += "</code></pre>"

	url, _ := cli.ServeGraphPage(page, info)
	return url, nil
}

func changesOptsStr(changesOpts changes.ChangesOptions) string {
	str := utils.AttrStr("since", changesOpts.SinceDate)
	str += utils.AttrStr("until", changesOpts.UntilDate)
	return str
}

func changesTimeseriesOptsStr(changesTimeseriesOpts changes.ChangesTimeseriesOptions) string {
	str := utils.AttrStr("since", changesTimeseriesOpts.Since)
	str += utils.AttrStr("until", changesTimeseriesOpts.Until)
	str += utils.AttrStr("period", changesTimeseriesOpts.Period)
	return str
}
