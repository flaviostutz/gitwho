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

func ExecGitBlame(repoPath string, filePath string, revision string) ([]BlameLine, error) {
	cmdResult, err := ExecShellf(repoPath, "/usr/bin/git blame --line-porcelain %s -- \"%s\"", revision, filePath)
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
		//         ![Conductor](docs/docs/img/conductor-vector-x.png) <<< TAB AT THE BEGINNING

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

func ExecPreviousCommitId(repoDir string, commitId string) (string, error) {
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git rev-list --parents -n 1 %s", commitId)
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

func ExecTreeFileSize(repoDir string, commitId string, filePath string) (int, error) {
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
	fmt.Printf("%v", parts)
	size, err := strconv.Atoi(parts[3])
	if err != nil {
		return -1, err
	}
	return size, nil
}

func ExecDiffTree(repoDir string, commitId string) ([]string, error) {
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git diff-tree --no-commit-id --name-only -r %s", commitId)
	if err != nil {
		return nil, err
	}
	lines, err := linesToArray(cmdResult)
	if err != nil {
		return nil, err
	}
	return lines, nil
}

func ExecCommitsInRange(repoDir string, branch string, since string, until string) ([]string, error) {
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git log --pretty=format:\"%%h\" --since=\"%s\" --until=\"%s\" %s", since, until, branch)
	if err != nil {
		return nil, err
	}
	if strings.Trim(cmdResult, "\n") == "" {
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

// FIXME check if this is used and delete it
func ExecGetCommitAtDate(repoDir string, branch string, when string) (string, error) {
	cmdResult, err := ExecShellf(repoDir, "/usr/bin/git rev-list -n 1 %s --until=\"%s\"", branch, when)
	if err != nil {
		return "", err
	}
	return strings.Trim(cmdResult, "\n"), nil
}
