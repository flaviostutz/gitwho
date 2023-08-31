package ownership

import (
	"fmt"
	"testing"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestTimelineOwnership1(t *testing.T) {
	logrus.SetLevel(logrus.DebugLevel)
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)
	results, err := TimelineCodeOwnership(OwnershipTimelineOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		MinDuplicateLines: 2,
		Until:             "now",
		Period:            "1 second",
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}

	require.Len(t, results, 2)
	require.NotEmpty(t, results[0].Commit.CommitId)
	require.NotEmpty(t, results[0].Commit.AuthorName)
	require.Equal(t, 3, results[0].TotalLines)
	require.Equal(t, 1, results[0].TotalFiles)
	require.Equal(t, 0, results[0].TotalLinesDuplicated)
	require.Equal(t, 2, len(results[0].AuthorsLines))

	require.Equal(t, 7, results[1].TotalLines)
	require.Equal(t, 2, results[1].TotalFiles)
	require.Equal(t, 0, results[1].TotalLinesDuplicated)
	require.Equal(t, 3, len(results[1].AuthorsLines))
}

func TestAnalyseCodeOwnershipAllFiles(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)

	commit, err := utils.ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)

	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		MinDuplicateLines: 2,
		CommitId:          commit.CommitId,
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}
	require.NotEmpty(t, results.Commit.CommitId)
	require.NotEmpty(t, results.Commit.AuthorName)
	require.Equal(t, 7, results.TotalLines)
	require.Equal(t, 2, results.TotalFiles)
	require.Equal(t, 0, results.TotalLinesDuplicated)
	require.Equal(t, 3, len(results.AuthorsLines))

	sumLines := 0
	for _, al := range results.AuthorsLines {
		sumLines += al.OwnedLinesTotal
	}
	require.Equal(t, results.TotalLines, sumLines)
}

func TestAnalyseCodeOwnershipAuthorRegex(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)

	commit, err := utils.ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)

	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir:      repoDir,
			Branch:       "main",
			AuthorsRegex: "author1",
		},
		MinDuplicateLines: 2,
		CommitId:          commit.CommitId,
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}
	require.NotEmpty(t, results.Commit.CommitId)
	require.NotEmpty(t, results.Commit.AuthorName)
	require.Equal(t, 1, results.TotalLines)
	require.Equal(t, 1, results.TotalFiles)
	require.Equal(t, 0, results.TotalLinesDuplicated)
	require.Equal(t, 1, len(results.AuthorsLines))

	sumLines := 0
	for _, al := range results.AuthorsLines {
		sumLines += al.OwnedLinesTotal
	}
	require.Equal(t, results.TotalLines, sumLines)
}

func TestAnalyseCodeOwnershipAuthorNotRegex(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)

	commit, err := utils.ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)

	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir:         repoDir,
			Branch:          "main",
			AuthorsNotRegex: "author2|author3",
		},
		MinDuplicateLines: 2,
		CommitId:          commit.CommitId,
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}
	require.NotEmpty(t, results.Commit.CommitId)
	require.NotEmpty(t, results.Commit.AuthorName)
	require.Equal(t, 1, results.TotalLines)
	require.Equal(t, 1, results.TotalFiles)
	require.Equal(t, 0, results.TotalLinesDuplicated)
	require.Equal(t, 1, len(results.AuthorsLines))

	sumLines := 0
	for _, al := range results.AuthorsLines {
		sumLines += al.OwnedLinesTotal
	}
	require.Equal(t, results.TotalLines, sumLines)
}

func TestAnalyseCodeOwnershipAuthorNotRegexMail(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)

	commit, err := utils.ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)

	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir:         repoDir,
			Branch:          "main",
			AuthorsNotRegex: "@mail.com",
		},
		MinDuplicateLines: 2,
		CommitId:          commit.CommitId,
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}
	require.NotEmpty(t, results.Commit.CommitId)
	require.NotEmpty(t, results.Commit.AuthorName)
	require.Equal(t, 0, results.TotalLines)
	require.Equal(t, 0, results.TotalFiles)
	require.Equal(t, 0, results.TotalLinesDuplicated)
	require.Equal(t, 0, len(results.AuthorsLines))
}

func TestAnalyseCodeOwnershipCheckSums(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipDuplicatesRepo()
	require.Nil(t, err)

	commit, err := utils.ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)

	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		MinDuplicateLines: 2,
		CommitId:          commit.CommitId,
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}

	sumLines := 0
	sumDup := 0
	sumDupOrig := 0
	sumDupOrigOthers := 0
	for _, al := range results.AuthorsLines {
		sumLines += al.OwnedLinesTotal
		sumDup += al.OwnedLinesDuplicate
		sumDupOrig += al.OwnedLinesDuplicateOriginal
		sumDupOrigOthers += al.OwnedLinesDuplicateOriginalOthers
	}
	fmt.Printf("%d\n", results.TotalLinesDuplicated)
	require.Equal(t, results.TotalLines, sumLines)
	require.Equal(t, results.TotalLinesDuplicated, sumDup)
	require.Equal(t, results.TotalLinesDuplicated, sumDupOrig+sumDupOrigOthers)
}

func TestAnalyseCodeDuplicates(t *testing.T) {
	repoDir, err := utils.ResolveTestOwnershipDuplicatesRepo()
	require.Nil(t, err)

	commit, err := utils.ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)

	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir: repoDir,
			Branch:  "main",
		},
		MinDuplicateLines: 2,
		CommitId:          commit.CommitId,
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}

	require.Equal(t, 2, results.TotalLinesDuplicated)
	require.Len(t, results.DuplicateLineGroups, 1)
	require.Len(t, results.DuplicateLineGroups[0].RelatedLinesGroup, 1)
	require.Equal(t, 2, results.DuplicateLineGroups[0].RelatedLinesCount)
	require.Equal(t, "file1", results.DuplicateLineGroups[0].FilePath)
	require.Equal(t, 2, results.DuplicateLineGroups[0].LineCount)
	require.Equal(t, "file2", results.DuplicateLineGroups[0].RelatedLinesGroup[0].FilePath)
	require.Equal(t, 2, results.DuplicateLineGroups[0].RelatedLinesGroup[0].LineCount)
}

func TestAnalyseCodeOwnershipRegexFiles(t *testing.T) {
	// require.InDeltaf(t, float64(0), v, 0.01, "")
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)

	commit, err := utils.ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)

	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir:    repoDir,
			Branch:     "main",
			FilesRegex: "/dir1.1/",
		},
		MinDuplicateLines: 2,
		CommitId:          commit.CommitId,
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}
	require.Equal(t, 5, results.TotalLines)
	require.Equal(t, 1, len(results.AuthorsLines))
}

func TestAnalyseCodeOwnershipRegexNotFiles(t *testing.T) {
	// require.InDeltaf(t, float64(0), v, 0.01, "")
	repoDir, err := utils.ResolveTestOwnershipRepo()
	require.Nil(t, err)

	commit, err := utils.ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)

	results, err := AnalyseCodeOwnership(OwnershipOptions{
		BaseOptions: utils.BaseOptions{
			RepoDir:       repoDir,
			Branch:        "main",
			FilesRegex:    "/dir1.1/",
			FilesNotRegex: "/dir1.1/",
		},
		MinDuplicateLines: 2,
		CommitId:          commit.CommitId,
	}, nil)
	require.Nil(t, err)
	if err != nil {
		return
	}
	require.Equal(t, 0, results.TotalLines)
	require.Equal(t, 0, len(results.AuthorsLines))
}
