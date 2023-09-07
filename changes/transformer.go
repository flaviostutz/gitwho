package changes

import (
	"fmt"
	"sort"
	"time"

	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
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

// ClusterizeAuthors processes changesResults and clusterizes authors using k-mean
func ClusterizeAuthors(changesResults []ChangesResult, numberOfClusters int) ([][]string, error) {
	linesAuthor := make(map[int][]string, 0)
	observations := make([]clusters.Observation, 0)

	authors := SortByAuthorDate(changesResults)
	for _, author := range authors {
		// count total touched
		authorTouched := 0
		for _, alines := range author.AuthorLinesDates {
			authorTouched += alines.AuthorLines.LinesTouched.New + alines.AuthorLines.LinesTouched.Changes
		}

		// prepare observations for clustering
		observations = append(observations, clusters.Coordinates{
			float64(authorTouched),
		})

		// index author by lines touched
		la, ok := linesAuthor[authorTouched]
		if !ok {
			la = make([]string, 0)
		}
		la = append(la, author.AuthorName)
		linesAuthor[authorTouched] = la
	}

	// clusterize authors
	km := kmeans.New()
	clusters, err := km.Partition(observations, numberOfClusters)
	if err != nil {
		return nil, err
	}

	// reverse clusters to author names
	authorClusters := make([][]string, 0)
	for _, c := range clusters {
		clusterAuthors := make([]string, 0)
		// fmt.Printf("Centered at x: %.2f\n", c.Center[0])
		// fmt.Printf("Matching data points: %+v\n\n", c.Observations)
		for _, obs := range c.Observations {
			totalTouched := int(obs.Coordinates()[0])
			authorNames, ok := linesAuthor[totalTouched]
			if !ok {
				fmt.Printf("SHOULDNT BE HERE!!!\n")
				panic(10)
			}
			clusterAuthors = append(clusterAuthors, authorNames...)
		}
		authorClusters = append(authorClusters, clusterAuthors)
	}

	return authorClusters, nil
}
