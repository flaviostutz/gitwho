package ownership

import (
	"fmt"
	"sort"
	"time"

	"github.com/muesli/clusters"
	"github.com/muesli/kmeans"
	"golang.org/x/exp/slices"
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

type AuthorCluster struct {
	Name    string
	Value   int
	Authors []string
}

type AuthorLinesCluster struct {
	Name        string
	Value       int
	AuthorLines []AuthorLines
}

// ClusterizeAuthors processes ownershipResults and clusterizes authors using k-mean over the number of owned lines
func ClusterizeAuthors(changesResults []OwnershipResult, numberOfClusters int) ([]AuthorLinesCluster, error) {
	linesAuthor := make(map[int][]string, 0)
	authorsLinesMap := make(map[string][]AuthorLines, 0)
	observations := make([]clusters.Observation, 0)

	authors := SortByAuthorDate(changesResults)
	for _, author := range authors {
		// map author by name
		authorLines, ok := authorsLinesMap[author.AuthorName]
		if !ok {
			authorLines = make([]AuthorLines, 0)
		}
		for _, ld := range author.AuthorLinesDate {
			authorLines = append(authorLines, ld.AuthorLines)
		}
		authorsLinesMap[author.AuthorName] = authorLines

		// count total owned
		authorOwned := 0
		for _, alines := range author.AuthorLinesDate {
			authorOwned += alines.AuthorLines.OwnedLinesTotal
		}

		// prepare observations for clustering
		observations = append(observations, clusters.Coordinates{
			float64(authorOwned),
		})

		// index author by lines touched
		la, ok := linesAuthor[authorOwned]
		if !ok {
			la = make([]string, 0)
		}
		la = append(la, author.AuthorName)
		linesAuthor[authorOwned] = la
	}

	// clusterize authors
	ncluster := numberOfClusters
	if len(observations) < ncluster {
		ncluster = len(observations)
	}

	km := kmeans.New()
	clusters, err := km.Partition(observations, ncluster)
	if err != nil {
		return nil, err
	}

	// reverse clusters to author names
	authorClusters := make([]AuthorCluster, 0)
	for _, c := range clusters {
		clusterAuthors := make([]string, 0)
		// fmt.Printf("Centered at x: %.2f\n", c.Center[0])
		// fmt.Printf("Matching data points: %+v\n\n", c.Observations)
		processedObs := make([]int, 0)
		for _, obs := range c.Observations {
			totalTouched := int(obs.Coordinates()[0])
			if slices.Contains(processedObs, totalTouched) {
				continue
			}
			authorNames := linesAuthor[totalTouched]
			clusterAuthors = append(clusterAuthors, authorNames...)
			processedObs = append(processedObs, totalTouched)
		}
		value := int(c.Center.Coordinates()[0])
		authorClusters = append(authorClusters, AuthorCluster{
			Name:    fmt.Sprintf("Authors that touched ~%d lines", value),
			Value:   value,
			Authors: clusterAuthors,
		})
	}

	// exchange author names by full author lines info
	authorLinesCluster := make([]AuthorLinesCluster, 0)
	for _, clusterWithName := range authorClusters {
		allAuthorLines := make([]AuthorLines, 0)
		for _, authorName := range clusterWithName.Authors {

			// sum author results
			authorLiness := authorsLinesMap[authorName]
			authorLines := AuthorLines{}
			for _, al := range authorLiness {
				authorLines.AuthorName = al.AuthorName
				authorLines.AuthorMail = al.AuthorMail
				authorLines.OwnedLinesAgeDaysSum += al.OwnedLinesAgeDaysSum
				authorLines.OwnedLinesDuplicate += al.OwnedLinesDuplicate
				authorLines.OwnedLinesDuplicateOriginal += al.OwnedLinesDuplicateOriginal
				authorLines.OwnedLinesDuplicateOriginalOthers += al.OwnedLinesDuplicateOriginalOthers
				authorLines.OwnedLinesTotal += al.OwnedLinesTotal
			}
			allAuthorLines = append(allAuthorLines, authorLines)
		}

		authorLinesCluster = append(authorLinesCluster, AuthorLinesCluster{
			Name:        clusterWithName.Name,
			Value:       clusterWithName.Value,
			AuthorLines: allAuthorLines,
		})
	}

	// order authors by lines owned count
	for _, authorLines := range authorLinesCluster {
		sort.Slice(authorLines.AuthorLines, func(i, j int) bool {
			return authorLines.AuthorLines[i].OwnedLinesTotal > authorLines.AuthorLines[j].OwnedLinesTotal
		})
	}

	// order clusters by lines owned count
	sort.Slice(authorLinesCluster, func(i, j int) bool {
		return authorLinesCluster[i].Value > (authorLinesCluster[j].Value)
	})

	return authorLinesCluster, nil
}
