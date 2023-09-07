package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type BlameLine struct {
	// AuthorName is the name of the last author that modified the line
	AuthorName string
	// AuthorMail is the mail of the last author that modified the line
	AuthorMail string
	// Date is when the original text of the line was introduced
	AuthorDate time.Time
	// Hash is the commit hash that introduced the original line
	CommitId     string
	LineContents string
}

type CommitInfo struct {
	AuthorName string    `json:"author_name"`
	AuthorMail string    `json:"author_mail"`
	Date       time.Time `json:"date"`
	CommitId   string    `json:"commit_id"`
}

func ExecGitBlame(repoPath string, filePath string, revision string) ([]BlameLine, error) {
	cmdResult, err := ExecShellf(repoPath, "/usr/bin/git blame --line-porcelain %s \"%s\"", revision, filePath)
	if err != nil {
		return nil, err
	}
	lines, err := linesToArray(cmdResult)
	if err != nil {
		return nil, err
	}

	blameLine := BlameLine{}
	result := make([]BlameLine, 0)
	next := true
	for _, line := range lines {
		// line-porcelain EXAMPLE
		// 92a77046dd4c181cc6810854c4dbe3a56e8ac6e9 1 1 1
		// author Viren Baraiya
		// author-mail <vbaraiya@netflix.com>
		// author-time 1484267213
		// author-tz -0800
		// committer Viren Baraiya
		// committer-mail <vbaraiya@netflix.com>
		// committer-time 1484267213
		// committer-tz -0800
		// summary updated URL for the image
		// previous 565949b9a9642ba7c79024c797c5d92876da5976 README.md <<< ONLY IF IT'S NOT NEW
		// filename README.md
		//         \![Conductor](docs/docs/img/conductor-vector-x.png) <<< TAB AT THE BEGINNING

		// start new line reading
		// fmt.Printf(">>>>>%s\n", line)
		if line == "" {
			continue
		}
		if next {
			next = false
			blameLine = BlameLine{CommitId: line[:40]}
		}

		if strings.HasPrefix(line, "author ") {
			blameLine.AuthorName = line[7:]
		}
		if strings.HasPrefix(line, "author-mail ") {
			blameLine.AuthorMail = line[12:]
		}
		if strings.HasPrefix(line, "author-time ") {
			epoch, err := strconv.ParseInt(line[12:], 10, 64)
			if err != nil {
				return nil, err
			}
			blameLine.AuthorDate = time.Unix(epoch, 0)
		}

		// end of line reading detected
		if strings.HasPrefix(line, "\t") {
			blameLine.LineContents = line[1:]
			result = append(result, blameLine)
			next = true
		}
	}
	return result, nil
}

func ExecListTree(repoDir string, commitId string) ([]string, error) {
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git ls-tree --name-only -r %s", commitId)
	if err != nil {
		return nil, err
	}
	lines, err := linesToArray(cmdResult)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

func ExecPreviousCommitIdForFile(repoDir string, commitId string, filePath string) (string, error) {
	cmdResult, err := ExecShellTimeout(repoDir, fmt.Sprintf("/usr/bin/git rev-list --parents -n 1 %s %s", commitId, filePath), 0, []int{0, 128})
	if err != nil {
		return "", err
	}
	cmdResult = strings.ReplaceAll(cmdResult, " ", "\n")
	lines, err := linesToArray(cmdResult)
	if err != nil {
		return "", err
	}
	if len(lines) != 2 {
		return "", nil
	}
	return lines[1], nil
}

func ExecDiffIsBinary(repoDir string, commitId string, filePath string) (bool, error) {
	// https://www.closedinterval.com/determine-if-a-file-is-binary-using-git/
	// fmt.Printf("/usr/bin/git diff 4b825dc642cb6eb9a060e54bf8d69288fbee4904 --numstat %s -- %s\n", commitId, filePath)
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git diff 4b825dc642cb6eb9a060e54bf8d69288fbee4904 --numstat %s -- %s", commitId, filePath)
	if err != nil {
		return false, err
	}
	if cmdResult == "" {
		return false, fmt.Errorf("No response returned")
	}
	isBinary := strings.HasPrefix(strings.ReplaceAll(cmdResult, "\t", ""), "--")
	return isBinary, nil
}

func ExecTreeFileSize(repoDir string, commitId string, filePath string) (int, error) {
	// fmt.Printf(">>> /usr/bin/git ls-tree -r --long %s %s", commitId, filePath)
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git ls-tree -r --long %s %s", commitId, filePath)
	if err != nil {
		return -1, err
	}
	cleanre := regexp.MustCompile(`\s+`)
	results := cleanre.ReplaceAllString(cmdResult, " ")
	if strings.Trim(results, " ") == "" {
		return -1, fmt.Errorf("File doesn't exist. commitId=%s; filePath=%s", commitId, filePath)
	}
	parts := strings.Split(results, " ")
	size, err := strconv.Atoi(parts[3])
	if err != nil {
		return -1, err
	}
	return size, nil
}

func ExecGitCommitInfo(repoDir string, commitId string) (CommitInfo, error) {
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git show -s --format=\"%%aN###<%%aE>---%%aI\" %s", commitId)
	if err != nil {
		return CommitInfo{}, err
	}
	if strings.Trim(cmdResult, " ") == "" {
		return CommitInfo{}, fmt.Errorf("Commit not found. commitId=%s", commitId)
	}
	result := strings.Trim(cmdResult, "\n")
	parts := strings.Split(result, "---")
	nparts := strings.Split(parts[0], "###")
	ctime, err := time.Parse(time.RFC3339, parts[1])
	if err != nil {
		return CommitInfo{}, err
	}
	return CommitInfo{Date: ctime, AuthorName: nparts[0], AuthorMail: nparts[1], CommitId: commitId}, nil
}

func ExecDiffTree(repoDir string, commitId1 string) ([]string, error) {
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git diff-tree --no-commit-id --root --name-only -r %s", commitId1)
	if err != nil {
		return nil, err
	}
	if strings.ReplaceAll(cmdResult, " ", "") == "" || strings.ReplaceAll(cmdResult, "\n", "") == "" {
		return []string{}, nil
	}
	lines, err := linesToArray(cmdResult)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

func ExecCommitIdsInDateRange(repoDir string, branch string, since string, until string) ([]string, error) {
	sinceStr := ""
	if since != "" {
		sinceStr = fmt.Sprintf("--since=\"%s\"", since)
	}
	untilStr := ""
	if until != "" {
		untilStr = fmt.Sprintf("--until=\"%s\"", until)
	}

	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git log --pretty=format:\"%%H\" %s %s %s", sinceStr, untilStr, branch)
	if err != nil {
		return nil, err
	}
	if strings.ReplaceAll(cmdResult, " ", "") == "" || strings.ReplaceAll(cmdResult, "\n", "") == "" {
		return []string{}, nil
	}
	commitIds, err := linesToArray(cmdResult)
	if err != nil {
		return nil, err
	}
	return commitIds, nil
}

func ExecCommitIdsInCommitRange(repoDir string, branch string, sinceCommit string, untilCommit string) ([]string, error) {
	commitRange := ""
	if sinceCommit != "" || untilCommit != "" {
		commitRange = fmt.Sprintf("%s...%s", sinceCommit, untilCommit)
	}
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git log --pretty=format:\"%%H\" %s %s", commitRange, branch)
	if err != nil {
		return nil, err
	}
	if strings.ReplaceAll(cmdResult, " ", "") == "" || strings.ReplaceAll(cmdResult, "\n", "") == "" {
		return []string{}, nil
	}
	commitIds, err := linesToArray(cmdResult)
	if err != nil {
		return nil, err
	}
	return commitIds, nil
}

func ExecDiffFileRevisions(repoDir string, filePath string, srcCommitId string, dstCommitId string) ([]DiffEntry, error) {
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git difftool -x \"diff\" --no-prompt %s %s -- \"%s\"", srcCommitId, dstCommitId, filePath)
	if err != nil {
		return nil, err
	}

	return ParseNormalDiffOutput(cmdResult)
}

func ExecGetCommitsInDateRange(repoDir string, branch string, since string, until string) ([]CommitInfo, error) {
	sinceStr := ""
	if since != "" {
		sinceStr = fmt.Sprintf("--since=\"%s\"", since)
	}
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git rev-list %s --until=\"%s\" --format=\"%%H---%%cI---%%cN---%%cE\" %s", sinceStr, until, branch)
	if err != nil {
		return nil, err
	}

	results, err := revListToCommitInfo(cmdResult)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func ExecGetCommitsInCommitRange(repoDir string, branch string, sinceCommit string, untilCommit string) ([]CommitInfo, error) {
	commitRange := ""
	if sinceCommit != "" || untilCommit != "" {
		commitRange = fmt.Sprintf("%s...%s", sinceCommit, untilCommit)
	}
	if commitRange == "" && branch == "" {
		return nil, fmt.Errorf("branch is required when sinceCommit and untilCommit are empty")
	}
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git rev-list %s %s --format=\"%%H---%%cI---%%cN---%%cE\"", branch, commitRange)
	if err != nil {
		return nil, err
	}

	results, err := revListToCommitInfo(cmdResult)
	if err != nil {
		return nil, err
	}

	return results, nil
}

func revListToCommitInfo(cmdResult string) ([]CommitInfo, error) {
	lines, err := linesToArray(cmdResult)
	if err != nil {
		return nil, err
	}

	results := make([]CommitInfo, 0)

	if len(lines) == 0 {
		return results, nil
	}

	for i := 1; i < len(lines); i += 2 {
		line := lines[i]
		parts := strings.Split(line, "---")

		date, err := time.Parse(time.RFC3339, parts[1])
		if err != nil {
			return nil, err
		}

		cinfo := CommitInfo{
			CommitId:   parts[0],
			Date:       date,
			AuthorName: parts[2],
			AuthorMail: parts[3],
		}
		results = append(results, cinfo)
	}
	return results, nil
}

func ExecGetLastestCommit(repoDir string, branch string, since string, until string) (*CommitInfo, error) {
	sinceStr := ""
	if since != "" {
		sinceStr = fmt.Sprintf("--since=\"%s\"", since)
	}
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git rev-list -n 1 %s --until=\"%s\" --format=\"%%H---%%cI---%%cN---%%cE\" %s", sinceStr, until, branch)
	if err != nil {
		return nil, err
	}

	result := strings.Split(cmdResult, "\n")
	if len(result) != 3 {
		return nil, nil
	}

	parts := strings.Split(result[1], "---")

	date, err := time.Parse(time.RFC3339, parts[1])
	if err != nil {
		return nil, err
	}

	return &CommitInfo{
		CommitId:   parts[0],
		Date:       date,
		AuthorName: parts[2],
		AuthorMail: parts[3],
	}, nil
}

func ExecCheckPrereqs() error {
	_, err := ExecShellf("", "/usr/bin/git version")
	if err != nil {
		return err
	}
	// TODO check git version
	return nil
}

func CommitInfoToCommitIds(cinfos []CommitInfo) []string {
	cids := make([]string, 0)
	for _, ci := range cinfos {
		cids = append(cids, ci.CommitId)
	}
	return cids
}
