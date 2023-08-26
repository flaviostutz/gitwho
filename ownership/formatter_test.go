package ownership

import (
	"testing"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/stretchr/testify/require"
)

func TestFormatOwnershipShort(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		When: "now",
	}, nil)
	require.Nil(t, err)

	out := FormatTextResults(results, false)
	require.Contains(t, out, "Total authors: 3\nTotal files: 2\nTotal lines: 7")
}

func TestFormatOwnershipFull(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		When: "now",
	}, nil)
	require.Nil(t, err)

	out := FormatTextResults(results, true)
	require.Contains(t, out, "Total authors: 3\nTotal files: 2\nAvg line age: 0 days\nDuplicated lines: 0")

}
