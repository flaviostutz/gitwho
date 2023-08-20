package utils

import (
	"regexp"
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
	FilePath   string
	LineNumber int
	AuthorName string
	AuthorMail string
	CommitDate time.Time
}

var (
	cleanRegex      = regexp.MustCompile("\t|\\s")
	ignoreLineRegex = regexp.MustCompile("^\\*|^\\/\\*|^#|import|from|package")
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
func (d *DuplicateLineTracker) AddLine(contents string, source LineSource) []LineSource {
	cline := cleanRegex.ReplaceAllString(contents, "")

	lines := strings.Split(cline, "\\n")
	for _, line := range lines {
		if len(line) < 15 || ignoreLineRegex.MatchString(line) {
			return nil
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
	lsources = append(lsources, source)
	d.lines[lineHash] = lsources
	d.mutex.Unlock()
	// fmt.Println(lineHash)
	return lsources
}
