package ownership

import (
	"testing"

	"github.com/flaviostutz/gitwho/ownership"
	"github.com/flaviostutz/gitwho/utils"
	"github.com/stretchr/testify/require"
)

func TestFormatCodeOwnershipShort(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)

	commit, err := utils.ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)

	results, err := ownership.AnalyseOwnership(ownership.OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		MinDuplicateLines: 2,
		CommitId:          commit.CommitId,
	}, nil)
	require.Nil(t, err)

	out, err := FormatCodeOwnershipResults(results, false)
	require.Nil(t, err)
	require.Contains(t, out, "Total authors: 3\nTotal files: 2\nTotal lines: 7")
}

func TestFormatCodeOwnershipFull(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)

	commit, err := utils.ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)

	results, err := ownership.AnalyseOwnership(ownership.OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		MinDuplicateLines: 2,
		CommitId:          commit.CommitId,
	}, nil)
	require.Nil(t, err)

	out, err := FormatCodeOwnershipResults(results, true)
	require.Nil(t, err)
	require.Contains(t, out, "Total authors: 3\nTotal files: 2\nAvg line age: 0 days\nDuplicated lines: 0")
}

func TestFormatDuplicatesFull(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)

	commit, err := utils.ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)

	results, err := ownership.AnalyseOwnership(ownership.OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		MinDuplicateLines: 2,
		CommitId:          commit.CommitId,
	}, nil)
	require.Nil(t, err)

	out := FormatDuplicatesResults(results, true)
	require.Contains(t, out, "Duplicated lines: 0 (0%)\n")
}
