package changes

import (
	"os"
	"testing"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/stretchr/testify/require"
)

var (
	sampleOpts = ChangesOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir:         "anypath/ANYREPODIR",
			Branch:          "feature/random-branch-here-12345",
			FilesRegex:      "/dir1.1/",
			FilesNotRegex:   "test",
			AuthorsRegex:    "abraham-neto|.*",
			AuthorsNotRegex: "Found.*There",
			CacheFile:       "gitwho-cache",
			CacheTTLSeconds: 10,
		},
		SinceDate: "12 months ago",
		UntilDate: "now",
	}

	sampleResult = ChangesResult{
		TotalFiles:        2123,
		TotalLinesTouched: LinesTouched{New: 123, Changes: 234, RefactorOwn: 111, RefactorOther: 222, RefactorReceived: 1, ChurnOwn: 444, ChurnOther: 555, ChurnReceived: 2, AgeDaysSum: 12333},
		TotalCommits:      123,
		AuthorsLines: []AuthorLines{
			{AuthorName: "author1", AuthorMail: "mail@mail.com", LinesTouched: LinesTouched{New: 123, Changes: 234, RefactorOwn: 111, RefactorOther: 222, RefactorReceived: 1, ChurnOwn: 444, ChurnOther: 555, ChurnReceived: 2, AgeDaysSum: 12333}, FilesTouched: []FileTouched{FileTouched{Name: "testeststs/tsetsetse", Lines: 123}}},
			{AuthorName: "author2", AuthorMail: "mail2@mail.com", LinesTouched: LinesTouched{New: 3444, Changes: 44, RefactorOwn: 222, RefactorOther: 222, RefactorReceived: 1, ChurnOwn: 444, ChurnOther: 2222, ChurnReceived: 2, AgeDaysSum: 12333}, FilesTouched: []FileTouched{FileTouched{Name: "aafafaf/sdfsdfds", Lines: 123}}},
		},
	}
)

func TestSaveNewCachedResultsOwnership(t *testing.T) {
	opts1 := sampleOpts // clone instance
	opts1.SinceDate = "123123123"
	os.Remove(opts1.CacheFile)

	result, err := GetFromCache(opts1)
	require.Nil(t, err)
	require.Nil(t, result)

	err = SaveToCache(opts1, sampleResult)
	require.Nil(t, err)

	result2, err := GetFromCache(opts1)
	require.Nil(t, err)
	require.NotNil(t, result2)
	require.Equal(t, sampleResult, *result2)
}
