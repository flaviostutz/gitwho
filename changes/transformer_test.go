package changes

import (
	"testing"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestClusterizeAuthors(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	logrus.SetLevel(logrus.DebugLevel)

	result, err := AnalyseTimeseriesChanges(ChangesTimeseriesOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
	}, nil)

	ClusterizeAuthors(result, 3)

}
