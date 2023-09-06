package changes

import (
	"sort"
	"time"
)

type AuthorNameLinesDate struct {
	AuthorName       string
	AuthorLinesDates []AuthorLinesDate
}

type AuthorLinesDate struct {
	Since       string
	AuthorLines AuthorLines
}

func SortByAuthorDate(changesResults []ChangesResult) []AuthorNameLinesDate {
	// map data
	authorNameDateLines := make(map[string]map[string]AuthorLines, 0)
	for _, result := range changesResults {
		date := result.SinceCommit.Date.Format(time.DateOnly)

		for _, authorLines := range result.AuthorsLines {
			authorDateLines, ok := authorNameDateLines[authorLines.AuthorName]
			if !ok {
				authorDateLines = make(map[string]AuthorLines, 0)
			}
			authorDateLines[date] = authorLines
			authorNameDateLines[authorLines.AuthorName] = authorDateLines
		}
	}

	// sort data
	authorNameLinesDates := make([]AuthorNameLinesDate, 0)

	for authorName, authorDateLines := range authorNameDateLines {
		authorLinesDates := make([]AuthorLinesDate, 0)
		for date, authorLines := range authorDateLines {
			authorChangesDate := AuthorLinesDate{
				Since:       date,
				AuthorLines: authorLines,
			}
			authorLinesDates = append(authorLinesDates, authorChangesDate)
		}
		authorNameLinesDate := AuthorNameLinesDate{
			AuthorName:       authorName,
			AuthorLinesDates: authorLinesDates,
		}
		authorNameLinesDates = append(authorNameLinesDates, authorNameLinesDate)
	}

	sort.Slice(authorNameLinesDates, func(i, j int) bool {
		return authorNameLinesDates[i].AuthorName < authorNameLinesDates[j].AuthorName
	})

	for _, authorData := range authorNameLinesDates {
		sort.Slice(authorData.AuthorLinesDates, func(i, j int) bool {
			return authorData.AuthorLinesDates[i].Since < authorData.AuthorLinesDates[j].Since
		})
	}

	return authorNameLinesDates
}
