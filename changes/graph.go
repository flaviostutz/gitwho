package changes

import (
	"github.com/flaviostutz/gitwho/utils"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

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
