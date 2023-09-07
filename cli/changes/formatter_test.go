package changes

import (
	"testing"

	"github.com/flaviostutz/gitwho/changes"
	"github.com/flaviostutz/gitwho/utils"
	"github.com/stretchr/testify/require"
)

func TestFormatChangesShort(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	results, err := changes.AnalyseChanges(changes.ChangesOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		SinceDate: "1 day ago",
	}, nil)
	require.Nil(t, err)

	out, err := FormatTopTextResults(results)
	require.Nil(t, err)
	require.Contains(t, out, "Top Coders (new+refactor-churn)\n  author3 <author3@mail.com>: 5")
}

func TestFormatChangesFull(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	results, err := changes.AnalyseChanges(changes.ChangesOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		SinceDate: "1 day ago",
	}, nil)
	require.Nil(t, err)

	out := FormatFullTextResults(results)
	require.Contains(t, out, "Total authors active: 3\nTotal files touched: 2\nAverage line age when changed: 0 days\n- Total lines touched: 11\n  - New lines: 8 (72%)\n  - Changed lines: 3 (27%)\n    - Refactor: 0 (0%)")

}
