package utils

import (
	"fmt"
	"strings"
)

func CalcPercStr(value int, total int) string {
	if total == 0 {
		return ""
	}
	return fmt.Sprintf(" (%d%%)", int(100*float64(value)/float64(total)))
}

func CalcDiffPercStr(curValue int, prevValue int) string {
	diff := curValue - prevValue
	if prevValue == 0 {
		return "       "
	}
	sig := ""
	if diff > 0 {
		sig = "+"
	}
	return fmt.Sprintf(" (%s%d%%)", sig, int(100*float64(diff)/float64(prevValue)))
}

func CalcDiffStr(curValue int, prevValue int) string {
	diff := curValue - prevValue
	if prevValue == 0 {
		return "       "
	}
	sig := ""
	if diff > 0 {
		sig = "+"
	}
	return fmt.Sprintf(" (%s%d)", sig, diff)
}

func BaseOptsStr(baseOpts BaseOptions) string {
	str := ""
	str += AttrStr("repo", baseOpts.RepoDir)
	str += AttrStr("branch", baseOpts.Branch)
	str += AttrStr("files", baseOpts.FilesRegex)
	str += AttrStr("files-not", baseOpts.FilesNotRegex)
	str += AttrStr("authors", baseOpts.AuthorsRegex)
	str += AttrStr("authors-not", baseOpts.AuthorsNotRegex)
	return str
}

func AttrStr(label string, value string) string {
	if value != "" {
		return fmt.Sprintf("%s: %s\n", label, value)
	}
	return ""
}

func JoinWithLimit(values []string, separator string, limit int) string {
	remaining := ""
	if len(values) > limit {
		remaining = fmt.Sprintf("... (+%d)", len(values)-limit)
		values = values[:limit]
	}
	valueStr := strings.Join(values, ", ")
	return fmt.Sprintf("%s%s", valueStr, remaining)
}
