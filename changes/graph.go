package changes

import (
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

// ServeChangesTimeseries Start server with a web page with graphs and
// returns the random URL generated for the page
func ServeChangesTimeseries(changesResults []ChangesResult, ownershipTimeseriesOpts ChangesTimeseriesOptions) string {

	// CHANGES TIMESERIES
	tr := charts.NewThemeRiver()
	tr.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine}),
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

	// TOTAL CHANGES PIE
	pie := charts.NewPie()

	authorTotals := make(map[string]int, 0)
	items := make([]opts.PieData, 0)
	for _, resultsTs := range changesResults {
		for _, authorLines := range resultsTs.AuthorsLines {
			if authorLines.LinesTouched.New+authorLines.LinesTouched.Changes == 0 {
				continue
			}
			authorTotal := authorTotals[authorLines.AuthorName]
			authorTotal += authorLines.LinesTouched.New + authorLines.LinesTouched.Changes
			authorTotals[authorLines.AuthorName] = authorTotal
		}
	}

	for authorName, authorTotal := range authorTotals {
		items = append(items, opts.PieData{Name: authorName, Value: authorTotal})
	}

	pie.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine}),
		charts.WithTitleOpts(opts.Title{
			Title: "Total Lines Touched",
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

	// ADD GRAPHS TO PAGE
	page := components.NewPage()
	page.SetLayout(components.PageFlexLayout)
	page.AddCharts(
		tr,
		pie,
	)

	info := "<pre style=\"display:flex;justify-content:center\"><code>"
	info += utils.BaseOptsStr(ownershipTimeseriesOpts.BaseOptions)
	info += changesTimeseriesOptsStr(ownershipTimeseriesOpts)
	info += FormatTimeseriesChangesResults(changesResults, true)
	info += "</code></pre>"

	url, _ := utils.ServeGraphPage(page, info)
	return url
}

// ServeChanges Start server with a web page with graphs and
// returns the random URL generated for the page
func ServeChanges(result ChangesResult, changesOpts ChangesOptions) string {

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
	links = append(links, opts.SankeyLink{Source: "Lines touched", Target: "Changed lines", Value: float32(result.TotalLinesTouched.Changes)})
	links = append(links, opts.SankeyLink{Source: "Changed lines", Target: "Refactor", Value: float32(result.TotalLinesTouched.RefactorOwn + result.TotalLinesTouched.RefactorOther)})
	links = append(links, opts.SankeyLink{Source: "Refactor", Target: "Refactor own", Value: float32(result.TotalLinesTouched.RefactorOwn)})
	links = append(links, opts.SankeyLink{Source: "Refactor", Target: "Refactor others", Value: float32(result.TotalLinesTouched.RefactorOther)})
	links = append(links, opts.SankeyLink{Source: "Changed lines", Target: "Churn", Value: float32(result.TotalLinesTouched.ChurnOwn + result.TotalLinesTouched.ChurnOther)})
	links = append(links, opts.SankeyLink{Source: "Churn", Target: "Churn own", Value: float32(result.TotalLinesTouched.ChurnOwn)})
	links = append(links, opts.SankeyLink{Source: "Churn", Target: "Churn others", Value: float32(result.TotalLinesTouched.ChurnOther)})
	links = append(links, opts.SankeyLink{Source: "Lines touched", Target: "New lines", Value: float32(result.TotalLinesTouched.New)})

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
	info += "</code></pre>"

	url, _ := utils.ServeGraphPage(page, info)
	return url
}

func changesOptsStr(changesOpts ChangesOptions) string {
	str := utils.AttrStr("since", changesOpts.Since)
	str += utils.AttrStr("until", changesOpts.Until)
	return str
}

func changesTimeseriesOptsStr(changesTimeseriesOpts ChangesTimeseriesOptions) string {
	str := utils.AttrStr("since", changesTimeseriesOpts.Since)
	str += utils.AttrStr("until", changesTimeseriesOpts.Until)
	str += utils.AttrStr("period", changesTimeseriesOpts.Period)
	return str
}
