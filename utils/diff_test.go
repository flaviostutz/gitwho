package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecDiffFiles(t *testing.T) {
	diffEntries, err := ExecDiffFiles("diff_file1.txt", "diff_file2.txt")
	assert.Nil(t, err)
	assert.NotEmpty(t, diffEntries)
	de := diffEntries[0]

	assert.Equal(t, OperationChange, de.Operation)
	assert.Equal(t, 1, de.DstLines[0].Number)
	assert.Equal(t, "First line CHANGED", de.DstLines[0].Text)
	assert.Equal(t, 1, de.SrcLines[0].Number)
	assert.Equal(t, 2, de.SrcLines[1].Number)
	assert.Equal(t, 3, de.SrcLines[2].Number)
	assert.Equal(t, "First line", de.SrcLines[0].Text)
	assert.Equal(t, "Second line", de.SrcLines[1].Text)
	assert.Equal(t, "Third line", de.SrcLines[2].Text)
}
