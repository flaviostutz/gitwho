package ownership

import (
	"fmt"
	"strconv"

	"github.com/flaviostutz/gitwho/ownership"
	"github.com/flaviostutz/gitwho/utils"
)

type authorLinesDate struct {
	date        string
	authorLines ownership.AuthorLines
}

func FormatCodeOwnershipResults(oresult ownership.OwnershipResult, full bool) (string, error) {
	// author clusters
	text := fmt.Sprintf("\nTotal authors: %d\n", len(oresult.AuthorsLines))
	text += fmt.Sprintf("Total files: %d\n", oresult.TotalFiles)

	if full {
		text += fmt.Sprintf("Avg line age: %s\n", avgLineAgeStr(oresult.LinesAgeDaysSum, oresult.TotalLines))
		text += fmt.Sprintf("Duplicated lines: %d (%d%%)\n", oresult.TotalLinesDuplicated, int(100*float64(oresult.TotalLinesDuplicated)/float64(oresult.TotalLines)))
	}

	// author clusters
	cstr, err := formatAuthorClusters([]ownership.OwnershipResult{oresult})
	if err != nil {
		return "", err
	}
	text += cstr

	text += fmt.Sprintf("Total lines: %d\n", oresult.TotalLines)
	for _, authorLines := range oresult.AuthorsLines {
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
			strconv.FormatFloat(float64(100)*(float64(authorLines.OwnedLinesTotal)/float64(oresult.TotalLines)), 'f', 1, 32),
			additional)
	}

	return text, nil
}

func FormatDuplicatesResults(ownershipResult ownership.OwnershipResult, full bool) string {
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
	return fmt.Sprintf("%1.f days", (linesAgeDaysSum / float64(totalLines)))
}

func formatAuthorClusters(cresult []ownership.OwnershipResult) (string, error) {
	aclusters, err := ownership.ClusterizeAuthors(cresult, 3)
	if err != nil {
		return "", fmt.Errorf("Couldn't clusterize authors. err=%s", err)
	}

	text := ""
	if len(aclusters) > 0 {
		text += "Author clusters:\n"
		for _, authorCluster := range aclusters {
			authorNames := make([]string, 0)
			for _, lines := range authorCluster.AuthorLines {
				authorNames = append(authorNames, lines.AuthorName)
			}
			text += fmt.Sprintf("  %s: %s\n", authorCluster.Name, utils.JoinWithLimit(authorNames, ", ", 4))
		}
	}

	return text, nil
}
