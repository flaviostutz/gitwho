package utils

import (
	"regexp"
	"time"

	"github.com/segmentio/fasthash/fnv1a"
)

// Attention: this utility will handle a lot of memory and a lot of calls
// Be careful about anything you are going to store and always think about optimization
type DuplicateLineTracker struct {
	// the key of the map is the hash of the contents of the line
	lines map[uint64][]LineSource
}

type LineSource struct {
	FilePath   string
	LineNumber string
	AuthorName string
	AuthorMail string
	CommitDate time.Time
}

var cleanRegex = regexp.MustCompile("\t|\\s|\n")

func NewDuplicateLineTracker() *DuplicateLineTracker {
	tracker := DuplicateLineTracker{}
	tracker.lines = make(map[uint64][]LineSource, 0)
	return &tracker
}

func (d *DuplicateLineTracker) AddLine(contents string, source LineSource) []LineSource {
	cline := cleanRegex.ReplaceAllString(contents, "")
	lineHash := fnv1a.HashString64(cline)
	lsources, ok := d.lines[lineHash]
	if !ok {
		d.lines[lineHash] = make([]LineSource, 1)
	}
	lsources = append(lsources, source)
	d.lines[lineHash] = lsources
	// fmt.Println(lineHash)
	return lsources
}
