package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecDiffFiles(t *testing.T) {
	diffEntries, err := ExecDiffFiles("diff_file1.txt", "diff_file2.txt")
	require.Nil(t, err)
	require.NotEmpty(t, diffEntries)
	de := diffEntries[0]

	require.Equal(t, OperationChange, de.Operation)
	require.Equal(t, 1, de.DstLines[0].Number)
	require.Equal(t, "First line CHANGED", de.DstLines[0].Text)
	require.Equal(t, 1, de.SrcLines[0].Number)
	require.Equal(t, 2, de.SrcLines[1].Number)
	require.Equal(t, 3, de.SrcLines[2].Number)
	require.Equal(t, "First line", de.SrcLines[0].Text)
	require.Equal(t, "Second line", de.SrcLines[1].Text)
	require.Equal(t, "Third line", de.SrcLines[2].Text)
}

// FIXME remove this later if not used. The diff quality is much more confusing than the diff tool
func TestDiffContents(t *testing.T) {
	diffs := DiffContents(
		`First line
Second line
Third line
Fourth line

Sixth line
Seventh line

Nineth line
`,
		`First line CHANGED
Fourth line

Sixth line
Additional line1
Additional line2
Seventh line
`)
	// for i, df := range diffs {
	// 	fmt.Printf("%d-%s\n%s\n", i+1, df.Type.String(), df.Text)
	// }
	require.Equal(t, 11, len(diffs))
}
