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

// ServeOwnershipTimeline Start server with a web page with graphs and
// returns the random URL generated for the page
func ServeOwnershipTimeline(ownershipResults []OwnershipResult, ownershipTimelineOpts OwnershipTimelineOptions) string {

	tr := charts.NewThemeRiver()
	tr.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeShine}),
		charts.WithTitleOpts(opts.Title{
			Title: "Ownership Timeline",
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

	page := components.NewPage()
	page.AddCharts(
		tr,
	)

	info := "<pre style=\"display:flex;justify-content:center\"><code>"
	info += utils.BaseOptsStr(ownershipTimelineOpts.BaseOptions)
	info += ownershipTimelineOptsStr(ownershipTimelineOpts)
	info += FormatTimelineOwnershipResults(ownershipResults, true)
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

func ownershipTimelineOptsStr(opts OwnershipTimelineOptions) string {
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
