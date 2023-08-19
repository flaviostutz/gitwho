package utils

import (
	"regexp"
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
	cleanRegex      = regexp.MustCompile("\t|\\s|\n")
	ignoreLineRegex = regexp.MustCompile("^\\*|^\\/\\*|^import")
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
func (d *DuplicateLineTracker) AddLine(contents string, source LineSource) []LineSource {
	cline := cleanRegex.ReplaceAllString(contents, "")

	if len(cline) < 30 || ignoreLineRegex.MatchString(cline) {
		return nil
	}

	lineHash := fnv1a.HashString64(cline)

	// this can make processing very slow because of thread syncronization
	// but is required for map access/change
	// TODO use an append-only strategy in a separate thread
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
