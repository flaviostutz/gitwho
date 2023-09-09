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
		date := result.UntilCommit.Date.Format(time.DateOnly)

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

type AuthorCluster struct {
	Name    string
	Value   int
	Authors []string
}

type AuthorLinesCluster struct {
	Name  string
	Value int
	Lines []AuthorLines
}

// ClusterizeAuthors processes changesResults and clusterizes authors using k-mean over the number of touched lines
func ClusterizeAuthors(changesResults []ChangesResult, numberOfClusters int) ([]AuthorLinesCluster, error) {
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
		for _, ld := range author.AuthorLinesDates {
			authorLines = append(authorLines, ld.AuthorLines)
		}
		authorsLinesMap[author.AuthorName] = authorLines

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
		for _, obs := range c.Observations {
			totalTouched := int(obs.Coordinates()[0])
			authorNames := linesAuthor[totalTouched]
			clusterAuthors = append(clusterAuthors, authorNames...)
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
			authorLines := AuthorLines{
				FilesTouched: make([]FileTouched, 0),
			}
			for _, al := range authorLiness {
				authorLines.AuthorName = al.AuthorName
				authorLines.AuthorMail = al.AuthorMail
				for _, ft := range authorLines.FilesTouched {
					authorLines.FilesTouched = append(authorLines.FilesTouched, ft)
				}
				authorLines.LinesTouched = SumLinesTouched(authorLines.LinesTouched, al.LinesTouched)
			}
			allAuthorLines = append(allAuthorLines, authorLines)
		}

		authorLinesCluster = append(authorLinesCluster, AuthorLinesCluster{
			Name:  clusterWithName.Name,
			Value: clusterWithName.Value,
			Lines: allAuthorLines,
		})
	}

	// order authors by lines touched count
	for _, authorLines := range authorLinesCluster {
		sort.Slice(authorLines.Lines, func(i, j int) bool {
			return (authorLines.Lines[i].LinesTouched.New + authorLines.Lines[i].LinesTouched.Changes) > (authorLines.Lines[j].LinesTouched.New + authorLines.Lines[j].LinesTouched.Changes)
		})
	}

	// order clusters by lines touched count
	sort.Slice(authorLinesCluster, func(i, j int) bool {
		return authorLinesCluster[i].Value > (authorLinesCluster[j].Value)
	})

	return authorLinesCluster, nil
}

func SumLinesTouched(changes1 LinesTouched, changes2 LinesTouched) LinesTouched {
	changes1.Changes += changes2.Changes
	changes1.ChurnOther += changes2.ChurnOther
	changes1.ChurnOwn += changes2.ChurnOwn
	changes1.ChurnReceived += changes2.ChurnReceived
	changes1.New += changes2.New
	changes1.RefactorOther += changes2.RefactorOther
	changes1.RefactorOwn += changes2.RefactorOwn
	changes1.RefactorReceived += changes2.RefactorReceived
	changes1.AgeDaysSum += changes2.AgeDaysSum
	return changes1
}
