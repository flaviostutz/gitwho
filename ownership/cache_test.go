package ownership

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/stretchr/testify/require"
)

var (
	sampleOpts = OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir:         "anypath/ANYREPODIR",
			Branch:          "feature/random-branch-here-12345",
			FilesRegex:      "/dir1.1/",
			FilesNotRegex:   "test",
			AuthorsRegex:    "abraham-neto|.*",
			AuthorsNotRegex: "Found.*There",
			CacheFile:       "gitwho.cache",
			CacheTTLSeconds: 10,
		},
		MinDuplicateLines: 4,
		CommitId:          "abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123",
	}

	sampleResult = OwnershipResult{
		TotalFiles: 2123,
		TotalLines: 354321,
		Commit: utils.CommitInfo{
			AuthorName: "Flavio de Oliveira Stutz",
			AuthorMail: "anmail_here@mail-mail.com",
			Date:       getSampleDate(),
			CommitId:   "abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123abc123",
		},
		TotalLinesDuplicated: 5463,
		LinesAgeDaysSum:      343.23,
		DuplicateLineGroups: []utils.LineGroup{
			{RelatedLinesCount: 123, Lines: utils.Lines{FilePath: "directory/file1.tst", LineNumber: 3, LineCount: 5}},
			{RelatedLinesCount: 13, Lines: utils.Lines{FilePath: "directory3/file6.tst", LineNumber: 88, LineCount: 9}},
		},
		AuthorsLines: []AuthorLines{
			{AuthorName: "author1", AuthorMail: "mail@mail.com", OwnedLinesTotal: 345, OwnedLinesAgeDaysSum: 23, OwnedLinesDuplicate: 222, OwnedLinesDuplicateOriginal: 12, OwnedLinesDuplicateOriginalOthers: 22},
			{AuthorName: "author2222", AuthorMail: "mail222@mail.com", OwnedLinesTotal: 111, OwnedLinesAgeDaysSum: 22, OwnedLinesDuplicate: 2122, OwnedLinesDuplicateOriginal: 122, OwnedLinesDuplicateOriginalOthers: 42},
		},
	}
)

func TestSaveNewCachedResultsOwnership(t *testing.T) {
	opts1 := sampleOpts // clone instance
	opts1.CommitId = "123123123"
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

func TestSaveExistingCachedResultsOwnership(t *testing.T) {
	opts1 := sampleOpts // clone instance
	opts1.CommitId = "abcabc"
	os.Remove(opts1.CacheFile)

	err := SaveToCache(opts1, sampleResult)
	require.Nil(t, err)

	err = SaveToCache(opts1, sampleResult)
	require.Nil(t, err)

	result2, err := GetFromCache(opts1)
	require.Nil(t, err)
	require.NotNil(t, result2)
	require.Equal(t, sampleResult, *result2)
}

func TestGetExistingCachedResultsOwnership(t *testing.T) {
	opts1 := sampleOpts // clone instance
	opts1.CommitId = "xyzxyz"
	os.Remove(opts1.CacheFile)

	err := SaveToCache(opts1, sampleResult)
	require.Nil(t, err)

	result2, err := GetFromCache(opts1)
	require.Nil(t, err)
	require.NotNil(t, result2)
	require.Equal(t, sampleResult, *result2)
}

func TestCacheAutoCleanup(t *testing.T) {
	opts1 := sampleOpts // clone instance
	opts1.CommitId = "bbbbbbbb"
	opts1.CacheTTLSeconds = 1
	os.Remove(opts1.CacheFile)

	err := SaveToCache(opts1, sampleResult)
	require.Nil(t, err)

	// not expired
	result2, err := GetFromCache(opts1)
	require.Nil(t, err)
	require.NotNil(t, result2)

	// wait for expiration
	time.Sleep(1100 * time.Millisecond)

	// expired
	result3, err := GetFromCache(opts1)
	require.Nil(t, err)
	require.Nil(t, result3)
}

func getSampleDate() time.Time {
	d, err := time.Parse(time.RFC3339, "2023-08-15T20:12:32-02:00")
	if err != nil {
		fmt.Println(err)
		panic(1)
	}
	return d
}
