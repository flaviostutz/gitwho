package ownership

import (
	"testing"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/stretchr/testify/require"
)

func TestClusterizeAuthors(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()

	require.Nil(t, err)
	results, err := AnalyseTimeseriesOwnership(OwnershipTimeseriesOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		MinDuplicateLines: 4,
		Since:             "2 days ago",
		Until:             "now",
		Period:            "1 second",
	}, nil)
	require.Nil(t, err)
	require.True(t, len(results) >= 2)

	authorClusters, err := ClusterizeAuthors(results, 2)
	require.Nil(t, err)
	require.Equal(t, 2, len(authorClusters))
	require.Equal(t, 1, len(authorClusters[0].AuthorLines))
	require.Equal(t, 2, len(authorClusters[1].AuthorLines))
	require.Equal(t, "author3", authorClusters[0].AuthorLines[0].AuthorName)
	require.Equal(t, "author1", authorClusters[1].AuthorLines[0].AuthorName)
	require.Equal(t, "author2", authorClusters[1].AuthorLines[1].AuthorName)
}
