package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNonDuplicateLines(t *testing.T) {
	dt := NewDuplicateLineTracker()

	lsources := dt.AddLine("abc", LineSource{})
	require.Len(t, lsources, 1)

	lsources = dt.AddLine("xyz", LineSource{})
	require.Len(t, lsources, 1)
}

func TestDuplicateLines(t *testing.T) {
	dt := NewDuplicateLineTracker()

	lsources := dt.AddLine("a  bc  dddddddddddddeeeeeeeefffffffggg     hhhhh iiiiii jjjjjjj kkkkkk       lmnoprs    tuv     xy       z   .", LineSource{})
	require.Len(t, lsources, 1)

	lsources = dt.AddLine("a b    	 	 c dddddddddddddeeeeeeeefffffffggg     hhhhh iiiiii jjjjjjj kkkkkk       lmnoprs    tuv     xy       z   .", LineSource{})
	require.Len(t, lsources, 2)
}

func TestDuplicateLineSourceOrder(t *testing.T) {
	dt := NewDuplicateLineTracker()

	lsources := dt.AddLine("abc", LineSource{AuthorName: "a"})
	lsources = dt.AddLine("abc", LineSource{AuthorName: "b"})
	lsources = dt.AddLine("abc", LineSource{AuthorName: "c"})

	require.Len(t, lsources, 3)
	require.Equal(t, "a", lsources[0].AuthorName)
	require.Equal(t, "b", lsources[1].AuthorName)
	require.Equal(t, "c", lsources[2].AuthorName)
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
