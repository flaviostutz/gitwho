package changes

import (
	"fmt"
	"testing"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestTimeseriesChanges1(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	repoDir, err := utils.ResolveTestOwnershipDuplicatesRepo()
	logs, _ := utils.ExecGetCommitsInDateRange(repoDir, "main", "", "")
	for _, l := range logs {
		fmt.Printf("%s %s\n", l.CommitId, l.Date.Format(time.RFC3339))
	}

	time.Sleep(1100 * time.Millisecond)

	require.Nil(t, err)
	results, err := AnalyseTimeseriesChanges(ChangesTimeseriesOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		Since:  "1 day",
		Until:  "now",
		Period: "1 second",
	}, nil)
	require.Nil(t, err)
	require.True(t, len(results) >= 2)
}

func TestAnalyseChangesNewFile2(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	logrus.SetLevel(logrus.DebugLevel)

	result, err := AnalyseChanges(ChangesOptions{
		BaseOptions: utils.BaseOptions{RepoDir: repoDir, Branch: "main", FilesRegex: "file2"},
	}, nil)

	// file2 was commited only one time with 5 lines of code

	require.Nil(t, err)
	require.Equal(t, 1, result.TotalCommits)
	require.Equal(t, 1, result.TotalFiles)
	require.Equal(t, 5, result.TotalLinesTouched.New)
	require.Equal(t, 0, result.TotalLinesTouched.Changes)

	require.Equal(t, 1, len(result.AuthorsLines))
	require.Equal(t, "author3", result.AuthorsLines[0].AuthorName)
	require.Equal(t, 1, len(result.AuthorsLines[0].FilesTouched))
	require.Equal(t, "dir1/dir1.1/file2", result.AuthorsLines[0].FilesTouched[0].Name)
	require.Equal(t, 5, result.AuthorsLines[0].FilesTouched[0].Lines)

}

func TestAnalyseChangesFile1(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	logrus.SetLevel(logrus.DebugLevel)

	result, err := AnalyseChanges(ChangesOptions{
		BaseOptions: utils.BaseOptions{RepoDir: repoDir, Branch: "main", FilesRegex: "file1"},
	}, nil)

	// file1 was commited 4 times

	require.Nil(t, err)
	require.Equal(t, 4, result.TotalCommits)
	require.Equal(t, 1, result.TotalFiles)
	require.Equal(t, 3, result.TotalLinesTouched.New)
	require.Equal(t, 3, result.TotalLinesTouched.Changes)

	require.Equal(t, 2, len(result.AuthorsLines))

	require.Equal(t, "author1", result.AuthorsLines[0].AuthorName)
	require.Equal(t, "file1", result.AuthorsLines[0].FilesTouched[0].Name)
	require.Equal(t, 4, result.AuthorsLines[0].FilesTouched[0].Lines)

	require.Equal(t, "author2", result.AuthorsLines[1].AuthorName)
	require.Equal(t, "file1", result.AuthorsLines[1].FilesTouched[0].Name)
	require.Equal(t, 2, result.AuthorsLines[1].FilesTouched[0].Lines)
}

func TestAnalyseChangesFile1Author1(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	logrus.SetLevel(logrus.DebugLevel)

	result, err := AnalyseChanges(ChangesOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir:      repoDir,
			Branch:       "main",
			FilesRegex:   "file1",
			AuthorsRegex: "author1",
		},
	}, nil)

	require.Nil(t, err)
	require.Equal(t, 3, result.TotalCommits)
	require.Equal(t, 1, result.TotalFiles)
	require.Equal(t, 2, result.TotalLinesTouched.New)
	require.Equal(t, 2, result.TotalLinesTouched.Changes)

	require.Equal(t, 1, len(result.AuthorsLines))

	require.Equal(t, "author1", result.AuthorsLines[0].AuthorName)
	require.Equal(t, "file1", result.AuthorsLines[0].FilesTouched[0].Name)
	require.Equal(t, 4, result.AuthorsLines[0].FilesTouched[0].Lines)
}

func TestAnalyseChangesNotAuthor1(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	logrus.SetLevel(logrus.DebugLevel)

	result, err := AnalyseChanges(ChangesOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir:         repoDir,
			Branch:          "main",
			FilesRegex:      "",
			AuthorsNotRegex: "author1",
		},
	}, nil)

	require.Nil(t, err)
	require.Equal(t, 2, result.TotalCommits)
	require.Equal(t, 2, result.TotalFiles)
	require.Equal(t, 6, result.TotalLinesTouched.New)
	require.Equal(t, 1, result.TotalLinesTouched.Changes)

	require.Equal(t, 2, len(result.AuthorsLines))

	require.Equal(t, "author3", result.AuthorsLines[0].AuthorName)
	require.Equal(t, "dir1/dir1.1/file2", result.AuthorsLines[0].FilesTouched[0].Name)
	require.Equal(t, 5, result.AuthorsLines[0].FilesTouched[0].Lines)

	require.Equal(t, "author2", result.AuthorsLines[1].AuthorName)
	require.Equal(t, "file1", result.AuthorsLines[1].FilesTouched[0].Name)
	require.Equal(t, 2, result.AuthorsLines[1].FilesTouched[0].Lines)
}

func TestAnalyseChangesAllFiles(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	logrus.SetLevel(logrus.DebugLevel)

	result, err := AnalyseChanges(ChangesOptions{
		BaseOptions: utils.BaseOptions{RepoDir: repoDir, Branch: "main", FilesRegex: "."},
	}, nil)

	require.Nil(t, err)
	require.Equal(t, 5, result.TotalCommits)
	require.Equal(t, 2, result.TotalFiles)
	require.Equal(t, 8, result.TotalLinesTouched.New)
	require.Equal(t, 3, result.TotalLinesTouched.Changes)

	require.Equal(t, 3, len(result.AuthorsLines))

	require.Equal(t, "author3", result.AuthorsLines[0].AuthorName)
	require.Equal(t, 1, len(result.AuthorsLines[0].FilesTouched))
	require.Equal(t, "dir1/dir1.1/file2", result.AuthorsLines[0].FilesTouched[0].Name)
	require.Equal(t, 5, result.AuthorsLines[0].FilesTouched[0].Lines)

	require.Equal(t, "author1", result.AuthorsLines[1].AuthorName)
	require.Equal(t, "file1", result.AuthorsLines[1].FilesTouched[0].Name)
	require.Equal(t, 4, result.AuthorsLines[1].FilesTouched[0].Lines)

	require.Equal(t, "author2", result.AuthorsLines[2].AuthorName)
	require.Equal(t, "file1", result.AuthorsLines[2].FilesTouched[0].Name)
	require.Equal(t, 2, result.AuthorsLines[2].FilesTouched[0].Lines)
}

func TestAnalyseChangesCheckTotals(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	logrus.SetLevel(logrus.DebugLevel)

	result, err := AnalyseChanges(ChangesOptions{
		BaseOptions: utils.BaseOptions{RepoDir: repoDir, Branch: "main", FilesRegex: "."},
	}, nil)

	rt := result.TotalLinesTouched
	require.Equal(t, rt.New+rt.Changes, rt.New+rt.ChurnOther+rt.ChurnOwn+rt.RefactorOther+rt.RefactorOwn)
	achanges := 0
	anew := 0
	alines := 0
	for _, authorLines := range result.AuthorsLines {
		achanges += authorLines.LinesTouched.Changes
		anew += authorLines.LinesTouched.New
		for _, authorFilesTouched := range authorLines.FilesTouched {
			alines += authorFilesTouched.Lines
		}
	}
	require.Equal(t, achanges, rt.Changes)
	require.Equal(t, anew, rt.New)
	require.Equal(t, alines, rt.New+rt.Changes)
}

func TestAnalyseChangesNotFiles(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	logrus.SetLevel(logrus.DebugLevel)

	result, err := AnalyseChanges(ChangesOptions{
		BaseOptions: utils.BaseOptions{RepoDir: repoDir, Branch: "main", FilesRegex: ".", FilesNotRegex: "dir1"},
	}, nil)

	require.Nil(t, err)
	require.Equal(t, 1, result.TotalFiles)
}

func TestCommitIdsForRange(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipDuplicatesRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}
	commitIds, sinceCommit, untilCommit, err := commitIdsForRange(ChangesOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
	})
	require.Nil(t, err)
	require.Equal(t, 4, len(commitIds))
	require.True(t, sinceCommit.Date.Before(untilCommit.Date))
}
