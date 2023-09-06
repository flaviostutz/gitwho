package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecGetLastestCommit(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	// should work for default master branch
	cid, err := ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)
	require.NotEmpty(t, cid.CommitId)
	require.Equal(t, "Your Name", cid.AuthorName)
	require.Equal(t, "you@example.com", cid.AuthorMail)

	// should fail for invalid branches
	cid, err = ExecGetLastestCommit(repoDir, "invalid-branch", "", "now")
	require.NotNil(t, err)
	require.Nil(t, cid)
}

func TestExecListTree(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	cid, err := ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)
	require.NotEmpty(t, cid)

	files, err := ExecListTree(repoDir, cid.CommitId)
	require.Nil(t, err)
	require.NotEmpty(t, files)

	require.Equal(t, 2, len(files))
}

func TestExecCommitDate(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	commitIds, err := ExecCommitIdsInRange(repoDir, "main", "1 month ago", "now")
	if err != nil {
		return
	}

	cinfo, err := ExecGitCommitInfo(repoDir, commitIds[0])
	require.Nil(t, err)
	require.False(t, cinfo.Date.IsZero())

	cinfo2, err := ExecGitCommitInfo(repoDir, commitIds[len(commitIds)-1])
	require.Nil(t, err)
	require.True(t, cinfo.Date.After(cinfo2.Date))
}

func TestExecDiffTree(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	cid, err := ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)
	require.NotEmpty(t, cid)

	files, err := ExecDiffTree(repoDir, cid.CommitId)
	require.Nil(t, err)
	require.NotEmpty(t, files)

	require.Equal(t, 1, len(files))
}

func TestExecTreeFileSize(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	commitIds, err := ExecCommitIdsInRange(repoDir, "main", "1 month ago", "now")
	if err != nil {
		return
	}

	size1, err := ExecTreeFileSize(repoDir, commitIds[0], "file1")
	require.Nil(t, err)
	require.Equal(t, 3, size1)

	size2, err := ExecTreeFileSize(repoDir, commitIds[len(commitIds)-1], "file1")
	require.Nil(t, err)
	require.Equal(t, 1, size2)

	require.NotEqual(t, size1, size2)
}

func TestExecDiffFileRevisions(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	commitIds, err := ExecCommitIdsInRange(repoDir, "main", "1 month ago", "now")
	if err != nil {
		return
	}

	de, err := ExecDiffFileRevisions(repoDir, "file1", commitIds[0], commitIds[len(commitIds)-1])
	require.Nil(t, err)
	require.Equal(t, OperationChange, de[0].Operation)
	require.Equal(t, 1, de[0].DstLines[0].Number)
	require.Equal(t, "a", de[0].DstLines[0].Text)
	require.Equal(t, 2, de[0].SrcLines[1].Number)
	require.Equal(t, "c", de[0].SrcLines[1].Text)
}

func TestExecCommitIdsInRange(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	cid, err := ExecCommitIdsInRange(repoDir, "main", "1 week ago", "now")
	require.Nil(t, err)
	require.NotEmpty(t, cid)
	require.Equal(t, 5, len(cid))
}

func TestExecCommitsInRange(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	commits, err := ExecGetCommitsInRange(repoDir, "main", "1 week ago", "now")
	require.Nil(t, err)
	require.Equal(t, 5, len(commits))
	require.NotEmpty(t, commits[0].AuthorName)
	require.NotEmpty(t, commits[0].CommitId)
}

func TestExecDiffIsBinary(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	cid, err := ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)
	require.NotEmpty(t, cid)

	isBin, err := ExecDiffIsBinary(repoDir, cid.CommitId, "file1")
	require.Nil(t, err)
	require.False(t, isBin)
}

func TestExecPreviousCommitIdForFile(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	cid, err := ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)
	require.NotEmpty(t, cid)

	prevCid, err := ExecPreviousCommitIdForFile(repoDir, cid.CommitId, "file1")
	require.Nil(t, err)
	require.NotEmpty(t, prevCid)
	require.NotEqual(t, prevCid, cid)
}

func TestExecGitBlame(t *testing.T) {
	repoDir, err := ResolveTestOwnershipRepo()
	require.Nil(t, err)
	if err != nil {
		return
	}

	cid, err := ExecGetLastestCommit(repoDir, "main", "", "now")
	require.Nil(t, err)
	require.NotEmpty(t, cid)

	lines, err := ExecGitBlame(repoDir, "file1", cid.CommitId)
	require.Nil(t, err)
	require.NotEmpty(t, lines)
	require.Equal(t, 2, len(lines))
	require.Equal(t, "author2", lines[0].AuthorName)
	require.Equal(t, "author1", lines[1].AuthorName)
}
