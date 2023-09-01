package ownership

import (
	"sort"
	"time"
)

type AuthorNameLinesDate struct {
	AuthorName      string
	AuthorLinesDate []AuthorLinesDate
}

type AuthorLinesDate struct {
	Date        string
	AuthorLines AuthorLines
}

func SortByAuthorDate(ownershipResults []OwnershipResult) []AuthorNameLinesDate {
	// map data
	authorNameDateLines := make(map[string]map[string]AuthorLines, 0)
	for _, result := range ownershipResults {
		date := result.Commit.Date.Format(time.DateOnly)

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
			authorLinesDate := AuthorLinesDate{
				Date:        date,
				AuthorLines: authorLines,
			}
			authorLinesDates = append(authorLinesDates, authorLinesDate)
		}
		authorNameLinesDate := AuthorNameLinesDate{
			AuthorName:      authorName,
			AuthorLinesDate: authorLinesDates,
		}
		authorNameLinesDates = append(authorNameLinesDates, authorNameLinesDate)
	}

	sort.Slice(authorNameLinesDates, func(i, j int) bool {
		return authorNameLinesDates[i].AuthorName < authorNameLinesDates[j].AuthorName
	})

	for _, authorData := range authorNameLinesDates {
		sort.Slice(authorData.AuthorLinesDate, func(i, j int) bool {
			return authorData.AuthorLinesDate[i].Date < authorData.AuthorLinesDate[j].Date
		})
	}

	return authorNameLinesDates
}
