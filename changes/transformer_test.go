package changes

import (
	"testing"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/stretchr/testify/require"
)

func TestClusterizeAuthors(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()

	require.Nil(t, err)
	results, err := AnalyseTimeseriesChanges(ChangesTimeseriesOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		Since:  "2 days ago",
		Until:  "now",
		Period: "1 second",
	}, nil)
	require.Nil(t, err)
	require.True(t, len(results) >= 2)

	authorClusters, err := ClusterizeAuthors(results, 3)
	require.Nil(t, err)
	require.Equal(t, 3, len(authorClusters))
	require.Equal(t, 1, len(authorClusters[0].AuthorLines))
	require.Equal(t, 1, len(authorClusters[1].AuthorLines))
	require.Equal(t, 1, len(authorClusters[2].AuthorLines))
	require.Equal(t, "author3", authorClusters[0].AuthorLines[0].AuthorName)
	require.NotEqual(t, authorClusters[1].AuthorLines[0].AuthorName, authorClusters[2].AuthorLines[0].AuthorName)
}
