package utils

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/fasthash/fnv1a"
)

// Attention: this utility will handle a lot of memory and a lot of calls
// Be careful about anything you are going to store and always think about optimization
type DuplicateLineTracker struct {
	// the key of the map is the hash of the contents of the line
	lines map[uint64][]LineSource
	mutex sync.Mutex
}

type LineSource struct {
	Lines
	AuthorName string
	AuthorMail string
	CommitDate time.Time
	lineHash   uint64
}

type LineGroup struct {
	Lines
	RelatedLinesGroup []LineGroup
	lineHashes        []uint64
}

type Lines struct {
	FilePath   string
	LineNumber int
	LineCount  int
}

var (
	cleanRegex      = regexp.MustCompile("\t|\\s")
	ignoreLineRegex = regexp.MustCompile("^\\*|^\\/\\*|^#|import|export|from|package")
)

func NewDuplicateLineTracker() *DuplicateLineTracker {
	tracker := DuplicateLineTracker{
		lines: make(map[uint64][]LineSource, 0),
		mutex: sync.Mutex{},
	}
	return &tracker
}

// Add a new line to tracker. If line is too short, it's is ignored and nil is returned
// This is thread safe, but can slow down parallelism in current implementation
// If string has string "\\n" (not \n), it will be split into distinct lines during ignore analysis
func (d *DuplicateLineTracker) AddLine(contents string, source LineSource) ([]LineSource, bool) {
	cline := cleanRegex.ReplaceAllString(contents, "")

	lines := strings.Split(cline, "\\n")
	for _, line := range lines {
		if len(line) < 15 || ignoreLineRegex.MatchString(line) {
			return nil, false
		}
	}

	lineHash := fnv1a.HashString64(cline)

	// this can slower processing because of thread syncronization
	// but is required for map access/change
	d.mutex.Lock()
	lsources, ok := d.lines[lineHash]
	if !ok {
		d.lines[lineHash] = make([]LineSource, 1)
	}
	source.lineHash = lineHash
	lsources = append(lsources, source)
	d.lines[lineHash] = lsources
	d.mutex.Unlock()
	isDuplicate := len(lsources) > 1
	// fmt.Println(lineHash)
	return lsources, isDuplicate
}

func (d *DuplicateLineTracker) GroupLines() []LineGroup {
	// lines map[uint64][]LineSource
	allLineSources := make([]LineSource, 0)
	allLineRefs := make([]uint64, 0)
	for key := range d.lines {
		lineSources := d.lines[key]
		for _, lineSource := range lineSources {
			allLineSources = append(allLineSources, lineSource)
			allLineRefs = append(allLineRefs, key)
		}
	}

	result := make([]LineGroup, 0)

	// form line groups from duplicated lines
	allLineGroups := findLineGroups(allLineSources)

	// for each line group, find line groups of the related lines to it
	foundRelatedLinesGroup := make(map[string]bool, 0)
	for i := range allLineGroups {
		lineGroup := allLineGroups[i]
		lineGroup.RelatedLinesGroup = make([]LineGroup, 0)
		lineGroupKey := groupKey(lineGroup)

		// gather all line sources related to this group
		lineSourcesForGroup := make([]LineSource, 0)
		for _, lineRef := range lineGroup.lineHashes {
			lineSourcesForGroup = append(lineSourcesForGroup, d.lines[lineRef]...)
		}

		// group line sources into groups
		candidateRelatedLineGroups := findLineGroups(lineSourcesForGroup)

		// add group as related to source line group
		for _, rlg := range candidateRelatedLineGroups {
			relatedGroupKey := groupKey(rlg)
			if lineGroupKey != relatedGroupKey {
				_, ok := foundRelatedLinesGroup[relatedGroupKey]
				if !ok {
					lineGroup.RelatedLinesGroup = append(lineGroup.RelatedLinesGroup, rlg)
				}
				foundRelatedLinesGroup[relatedGroupKey] = true
			}
		}
		_, ok := foundRelatedLinesGroup[lineGroupKey]
		if !ok {
			result = append(result, lineGroup)
			foundRelatedLinesGroup[lineGroupKey] = true
		}
	}

	return result
}

func groupKey(lg LineGroup) string {
	return fmt.Sprintf("%s#%d#%d", lg.FilePath, lg.LineCount, lg.LineNumber)
}

func findLineGroups(lineSources []LineSource) []LineGroup {
	// this allLineSources list is ordered by fileName and line number
	// this is the basis for this algorithm to work
	sort.Slice(lineSources, func(i, j int) bool {
		return lineSources[i].FilePath < lineSources[j].FilePath &&
			lineSources[i].LineNumber < lineSources[j].LineNumber
	})

	lineGroups := make([]LineGroup, 0)
	currentDup := LineGroup{}
	for _, lineSource := range lineSources {
		if currentDup.LineNumber == 0 {
			currentDup.FilePath = lineSource.FilePath
			currentDup.LineNumber = lineSource.LineNumber
			currentDup.LineCount = lineSource.LineCount
		}

		overlap, lineNumber, lineCount := mergeOverlap(currentDup.Lines, lineSource.Lines)

		if overlap {
			currentDup.LineNumber = lineNumber
			currentDup.LineCount = lineCount
			currentDup.lineHashes = append(currentDup.lineHashes, uint64(lineSource.lineHash))
			continue
		}

		// found lines that don't overlap. add current dup and start a new duplicate instance
		lineGroups = append(lineGroups, currentDup)
		currentDup = LineGroup{}
		currentDup.FilePath = lineSource.FilePath
		currentDup.LineNumber = lineSource.LineNumber
		currentDup.LineCount = lineSource.LineCount
	}
	if currentDup.LineNumber != 0 {
		lineGroups = append(lineGroups, currentDup)
	}
	return lineGroups
}

func mergeOverlap(lines1 Lines, lines2 Lines) (bool, int, int) {
	if lines1.FilePath != lines2.FilePath {
		return false, -1, -1
	}
	from1 := lines1.LineNumber
	to1 := lines1.LineNumber + lines1.LineCount
	from2 := lines2.LineNumber
	to2 := lines2.LineNumber + lines2.LineCount

	minFrom := from1
	if from2 < from1 {
		minFrom = from2
	}

	maxTo := to2
	if to1 > to2 {
		maxTo = to1
	}

	for i := minFrom; i <= maxTo; i++ {
		if i > to1 && i < from2 {
			return false, -1, -1
		}
		if i > to2 && i < from1 {
			return false, -1, -1
		}
	}

	return true, minFrom, maxTo - minFrom
}
