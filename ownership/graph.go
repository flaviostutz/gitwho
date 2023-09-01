package ownership

import (
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// ServeOwnershipTimeline Start server with a web page with graphs and
// returns the random URL generated for the page
func ServeOwnershipTimeline(ownershipResults []OwnershipResult) string {

	tr := charts.NewThemeRiver()
	tr.SetGlobalOptions(
		charts.WithSingleAxisOpts(opts.SingleAxis{
			Type:   "time",
			Bottom: "10%",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Trigger: "axis",
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

	tr.AddSeries("themeRiver", data)

	page := components.NewPage()
	page.AddCharts(
		tr,
	)

	url, _ := utils.ServeGraphPage(page, "")
	return url
}

func ServeOwnership(ownershipResult OwnershipResult) string {
	pie := charts.NewPie()

	items := make([]opts.PieData, 0)
	for _, authorLines := range ownershipResult.AuthorsLines {
		items = append(items, opts.PieData{Name: authorLines.AuthorName, Value: authorLines.OwnedLinesTotal})
	}
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

	contents := "<pre style=\"display:flex;justify-content:center\"><code>"
	contents += FormatCodeOwnershipResults(ownershipResult, true)
	contents += "</code></pre>"

	url, _ := utils.ServeGraphPage(page, contents)
	return url
}
