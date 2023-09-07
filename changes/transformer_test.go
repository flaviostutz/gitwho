package changes

import (
	"testing"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/stretchr/testify/require"
)

func TestClusterizeAuthors(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipDuplicatesRepo()

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

	authorClusters, err := ClusterizeAuthors(results, 2)
	require.Nil(t, err)
	require.Equal(t, 2, len(authorClusters))
}
