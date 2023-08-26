package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNonDuplicateLines(t *testing.T) {
	dt := NewDuplicateLineTracker()

	lsources, dup := dt.AddLine("abc01234567890123456789", LineSource{})
	require.False(t, dup)
	require.Len(t, lsources, 1)

	lsources, dup = dt.AddLine("xyz01234567890123456789", LineSource{})
	require.False(t, dup)
	require.Len(t, lsources, 1)
}

func TestDuplicateLines(t *testing.T) {
	dt := NewDuplicateLineTracker()

	lsources, dup := dt.AddLine("a  bc  dddddddddddddeeeeeeeefffffffggg     hhhhh iiiiii jjjjjjj kkkkkk       lmnoprs    tuv     xy       z   .", LineSource{})
	require.False(t, dup)
	require.Len(t, lsources, 1)

	lsources, dup = dt.AddLine("a b    	 	 c dddddddddddddeeeeeeeefffffffggg     hhhhh iiiiii jjjjjjj kkkkkk       lmnoprs    tuv     xy       z   .", LineSource{})
	require.True(t, dup)
	require.Len(t, lsources, 2)
}

func TestDuplicateLineSourceOrder(t *testing.T) {
	dt := NewDuplicateLineTracker()

	lsources, dup := dt.AddLine("abc01234567890123456789", LineSource{AuthorName: "a"})
	require.False(t, dup)
	lsources, dup = dt.AddLine("abc01234567890123456789", LineSource{AuthorName: "b"})
	require.True(t, dup)
	lsources, dup = dt.AddLine("abc01234567890123456789", LineSource{AuthorName: "c"})
	require.True(t, dup)

	require.Len(t, lsources, 3)
	require.Equal(t, "a", lsources[0].AuthorName)
	require.Equal(t, "b", lsources[1].AuthorName)
	require.Equal(t, "c", lsources[2].AuthorName)
}

func TestDuplicateLineGroups(t *testing.T) {
	dt := NewDuplicateLineTracker()

	dt.AddLine("abc01234567890123456789", LineSource{Lines: Lines{FilePath: "2", LineNumber: 20, LineCount: 1}, AuthorName: "b"})
	dt.AddLine("abc01234567890123456789", LineSource{Lines: Lines{FilePath: "3", LineNumber: 30, LineCount: 1}, AuthorName: "c"})
	dt.AddLine("abc01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 10, LineCount: 1}, AuthorName: "a"})

	lineGroups := dt.GroupLines()
	require.Len(t, lineGroups, 1)
	require.Len(t, lineGroups[0].RelatedLinesGroup, 2)

	require.Equal(t, "1", lineGroups[0].FilePath)
	require.Equal(t, 10, lineGroups[0].LineNumber)
	require.Equal(t, 1, lineGroups[0].LineCount)
	require.Equal(t, "2", lineGroups[0].RelatedLinesGroup[0].FilePath)
	require.Equal(t, "3", lineGroups[0].RelatedLinesGroup[1].FilePath)
}

func TestDuplicateLineGroups2(t *testing.T) {
	dt := NewDuplicateLineTracker()

	dt.AddLine("abc01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 10, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("xyz01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 11, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("abc01234567890123456789", LineSource{Lines: Lines{FilePath: "2", LineNumber: 20, LineCount: 1}, AuthorName: "b"})
	dt.AddLine("abc01234567890123456789", LineSource{Lines: Lines{FilePath: "3", LineNumber: 30, LineCount: 1}, AuthorName: "c"})

	lineGroups := dt.GroupLines()
	require.Len(t, lineGroups, 1)
	require.Len(t, lineGroups[0].RelatedLinesGroup, 2)

	require.Equal(t, "1", lineGroups[0].FilePath)
	require.Equal(t, 10, lineGroups[0].LineNumber)
	require.Equal(t, 2, lineGroups[0].LineCount)
	require.Equal(t, "2", lineGroups[0].RelatedLinesGroup[0].FilePath)
	require.Equal(t, "3", lineGroups[0].RelatedLinesGroup[1].FilePath)
}

func TestDuplicateLineGroups3(t *testing.T) {
	dt := NewDuplicateLineTracker()

	dt.AddLine("abc01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 10, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("xyz01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 11, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("abc01234567890123456789", LineSource{Lines: Lines{FilePath: "2", LineNumber: 20, LineCount: 1}, AuthorName: "b"})
	dt.AddLine("abc01234567890123456789", LineSource{Lines: Lines{FilePath: "3", LineNumber: 30, LineCount: 1}, AuthorName: "c"})
	dt.AddLine("xyz01234567890123456789", LineSource{Lines: Lines{FilePath: "3", LineNumber: 40, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("abc01234567890123456789", LineSource{Lines: Lines{FilePath: "4", LineNumber: 50, LineCount: 1}, AuthorName: "d"})
	dt.AddLine("xyz01234567890123456789", LineSource{Lines: Lines{FilePath: "4", LineNumber: 51, LineCount: 1}, AuthorName: "d"})

	lineGroups := dt.GroupLines()
	require.Len(t, lineGroups, 1)
	require.Len(t, lineGroups[0].RelatedLinesGroup, 4)

	require.Equal(t, "1", lineGroups[0].FilePath)
	require.Equal(t, 10, lineGroups[0].LineNumber)
	require.Equal(t, 2, lineGroups[0].LineCount)
	require.Equal(t, "2", lineGroups[0].RelatedLinesGroup[0].FilePath)
	require.Equal(t, "3", lineGroups[0].RelatedLinesGroup[1].FilePath)
	require.Equal(t, "3", lineGroups[0].RelatedLinesGroup[2].FilePath)
	require.Equal(t, "4", lineGroups[0].RelatedLinesGroup[3].FilePath)
}

func TestFindLineGroups1(t *testing.T) {
	lines := make([]LineSource, 0)
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 1, LineCount: 1}, lineHash: 1})
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 2, LineCount: 2}, lineHash: 2})
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 5, LineCount: 4}, lineHash: 1})
	lineGroups := findLineGroups(lines)
	require.Len(t, lineGroups, 2)
	require.Equal(t, 1, lineGroups[0].LineNumber)
	require.Equal(t, 3, lineGroups[0].LineCount)
	require.Equal(t, 5, lineGroups[1].LineNumber)
	require.Equal(t, 4, lineGroups[1].LineCount)
}

func TestFindLineGroupsNoOrder1(t *testing.T) {
	lines := make([]LineSource, 0)
	// line appearance in reverse ordem should work
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 1, LineCount: 1}, lineHash: 1})
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 5, LineCount: 4}, lineHash: 1})
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 2, LineCount: 2}, lineHash: 2})
	lineGroups := findLineGroups(lines)
	require.Len(t, lineGroups, 2)
	require.Equal(t, 1, lineGroups[0].LineNumber)
	require.Equal(t, 3, lineGroups[0].LineCount)
	require.Equal(t, 5, lineGroups[1].LineNumber)
	require.Equal(t, 4, lineGroups[1].LineCount)
}

func TestFindLineGroups2(t *testing.T) {
	lines := make([]LineSource, 0)
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 1, LineCount: 1}, lineHash: 1})
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 2, LineCount: 2}, lineHash: 2})
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 4, LineCount: 4}, lineHash: 1})
	lineGroups := findLineGroups(lines)
	require.Len(t, lineGroups, 1)
	require.Equal(t, 1, lineGroups[0].LineNumber)
	require.Equal(t, 7, lineGroups[0].LineCount)
}

func TestFindLineGroups3(t *testing.T) {
	lines := make([]LineSource, 0)
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 1, LineCount: 1}, lineHash: 1})
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 2, LineCount: 2}, lineHash: 2})
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 4, LineCount: 4}, lineHash: 1})
	lines = append(lines, LineSource{Lines: Lines{FilePath: "1", LineNumber: 4, LineCount: 4}, lineHash: 3})
	lineGroups := findLineGroups(lines)
	require.Len(t, lineGroups, 1)
	require.Equal(t, 1, lineGroups[0].LineNumber)
	require.Equal(t, 7, lineGroups[0].LineCount)
}

func TestDuplicateLineGroups4(t *testing.T) {
	dt := NewDuplicateLineTracker()

	dt.AddLine("aaa01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 10, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("bbb01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 11, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("ccc01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 12, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("ddd01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 13, LineCount: 1}, AuthorName: "a"})

	dt.AddLine("aaa01234567890123456789", LineSource{Lines: Lines{FilePath: "2", LineNumber: 210, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("bbb01234567890123456789", LineSource{Lines: Lines{FilePath: "2", LineNumber: 211, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("ccc01234567890123456789", LineSource{Lines: Lines{FilePath: "2", LineNumber: 212, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("ddd01234567890123456789", LineSource{Lines: Lines{FilePath: "2", LineNumber: 213, LineCount: 1}, AuthorName: "a"})

	dt.AddLine("aaa01234567890123456789", LineSource{Lines: Lines{FilePath: "3", LineNumber: 110, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("bbb01234567890123456789", LineSource{Lines: Lines{FilePath: "3", LineNumber: 111, LineCount: 1}, AuthorName: "a"})

	lineGroups := dt.GroupLines()
	require.Len(t, lineGroups, 1)
	require.Len(t, lineGroups[0].RelatedLinesGroup, 2)

	require.Equal(t, "1", lineGroups[0].FilePath)
	require.Equal(t, 10, lineGroups[0].LineNumber)
	require.Equal(t, 4, lineGroups[0].LineCount)
	require.Equal(t, "2", lineGroups[0].RelatedLinesGroup[0].FilePath)
	require.Equal(t, "3", lineGroups[0].RelatedLinesGroup[1].FilePath)
}

func TestDuplicateLineGroupsOutOfOrder4(t *testing.T) {
	dt := NewDuplicateLineTracker()

	dt.AddLine("bbb01234567890123456789", LineSource{Lines: Lines{FilePath: "3", LineNumber: 111, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("bbb01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 11, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("aaa01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 10, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("ccc01234567890123456789", LineSource{Lines: Lines{FilePath: "2", LineNumber: 212, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("ccc01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 12, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("ddd01234567890123456789", LineSource{Lines: Lines{FilePath: "1", LineNumber: 13, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("aaa01234567890123456789", LineSource{Lines: Lines{FilePath: "3", LineNumber: 110, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("ddd01234567890123456789", LineSource{Lines: Lines{FilePath: "2", LineNumber: 213, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("bbb01234567890123456789", LineSource{Lines: Lines{FilePath: "2", LineNumber: 211, LineCount: 1}, AuthorName: "a"})
	dt.AddLine("aaa01234567890123456789", LineSource{Lines: Lines{FilePath: "2", LineNumber: 210, LineCount: 1}, AuthorName: "a"})

	lineGroups := dt.GroupLines()
	require.Len(t, lineGroups, 1)
	require.Len(t, lineGroups[0].RelatedLinesGroup, 2)

	require.Equal(t, "1", lineGroups[0].FilePath)
	require.Equal(t, 10, lineGroups[0].LineNumber)
	require.Equal(t, 4, lineGroups[0].LineCount)
	require.Equal(t, "2", lineGroups[0].RelatedLinesGroup[0].FilePath)
	require.Equal(t, "3", lineGroups[0].RelatedLinesGroup[1].FilePath)
}

func BenchmarkNonDuplicatedLines(b *testing.B) {
	dt := NewDuplicateLineTracker()

	for i := 0; i < b.N; i++ {
		dt.AddLine(fmt.Sprintf("0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789%d", i), LineSource{})
	}

	require.Len(b, dt.lines, b.N)
}

func BenchmarkDuplicateLines(b *testing.B) {
	dt := NewDuplicateLineTracker()

	for i := 0; i < b.N; i++ {
		// using fmt.Sprintf here so the testing function cost is similar to BenchmarkDifferentLines
		dt.AddLine(fmt.Sprintf("012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678%d", 9), LineSource{})
	}

	require.Len(b, dt.lines, 1)
}

func BenchmarkMixedDuplicatedLines(b *testing.B) {
	dt := NewDuplicateLineTracker()

	for i := 0; i < b.N; i++ {
		// 10% of tested lines will be different
		value := 1
		if b.N >= 10 {
			value = i % (b.N / 10)
		}
		dt.AddLine(fmt.Sprintf("0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789%d", value), LineSource{})
	}

	if b.N >= 10 {
		require.Len(b, dt.lines, b.N/10)
	} else {
		require.Len(b, dt.lines, b.N)
	}
}
