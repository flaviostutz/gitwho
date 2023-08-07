package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type DiffOp int

const (
	OperationAdd DiffOp = iota
	OperationDelete
	OperationChange
)

type LineText struct {
	Number int
	Text   string
}

type DiffEntry struct {
	Operation DiffOp
	SrcLines  []LineText
	DstLines  []LineText
}

var (
	srcTextRe = regexp.MustCompile("^<\\s(.*)$")
	dstTextRe = regexp.MustCompile("^>\\s(.*)$")
	newOpRe   = regexp.MustCompile("^(\\d*,{0,1}\\d*)(a|d|c)(\\d*,{0,1}\\d*)$")
)

func ExecDiffFiles(fileSrc string, fileDst string) ([]DiffEntry, error) {
	cmdResult, err := ExecShellTimeout("", fmt.Sprintf("/usr/bin/diff \"%s\" \"%s\"", fileSrc, fileDst), 0, 1)
	if err != nil {
		return nil, err
	}

	lines, err := linesToArray(cmdResult)
	if err != nil {
		return nil, err
	}

	// parse results

	result := make([]DiffEntry, 0)
	li := 0
	for li < len(lines) {
		// diff output EXAMPLE
		// 1,3c1
		// < First line
		// < Second line
		// < Third line
		// ---
		// > First line CHANGED
		// 6a5,6
		// > Additional line1
		// > Additional line2
		// 8,9d7
		// <
		// < Nineth line

		line := lines[li]

		// diff op
		newOpMatches := newOpRe.FindStringSubmatch(line)
		if newOpMatches == nil {
			return nil, fmt.Errorf("Couldn't extract diff operation. line=%s", line)
		}

		diffEntry := DiffEntry{}

		// operation extraction
		switch newOpMatches[2] {
		case "a":
			diffEntry.Operation = OperationAdd
		case "d":
			diffEntry.Operation = OperationDelete
		case "c":
			diffEntry.Operation = OperationChange
		default:
			err := fmt.Errorf("Invalid operation received. op=%s", newOpMatches[2])
			return nil, err
		}

		// src line extraction
		srcLines, err := extractLines(newOpMatches[1])
		if err != nil {
			return nil, err
		}
		diffEntry.SrcLines = srcLines

		// dst line extraction
		dstLines, err := extractLines(newOpMatches[3])
		if err != nil {
			return nil, err
		}
		diffEntry.DstLines = dstLines

		// src text extraction
		if diffEntry.Operation == OperationDelete || diffEntry.Operation == OperationChange {
			for i := 0; i < len(diffEntry.SrcLines); i++ {
				li += 1
				line = lines[li]
				srcTextMatches := srcTextRe.FindStringSubmatch(line)
				if srcTextMatches == nil {
					return nil, fmt.Errorf("Cannot find src text. line=%s", line)
				}
				diffEntry.SrcLines[i].Text = srcTextMatches[1]
			}
		}

		if diffEntry.Operation == OperationChange {
			// skip separator line "---"
			li += 1
		}

		// dst text extraction
		if diffEntry.Operation == OperationAdd || diffEntry.Operation == OperationChange {
			for i := 0; i < len(diffEntry.DstLines); i++ {
				li += 1
				line = lines[li]
				dstTextMatches := dstTextRe.FindStringSubmatch(line)
				if dstTextMatches == nil {
					return nil, fmt.Errorf("Cannot find dst text. line=%s", line)
				}
				diffEntry.DstLines[i].Text = dstTextMatches[1]
			}
		}

		result = append(result, diffEntry)
		li += 1
	}

	return result, nil
}

func extractLines(input string) ([]LineText, error) {
	elems := strings.Split(input, ",")

	line1, err := strconv.Atoi(elems[0])
	if err != nil {
		return nil, err
	}
	line2 := line1

	// has a range of dst lines
	if len(elems) == 2 {
		line2, err = strconv.Atoi(elems[1])
		if err != nil {
			return nil, err
		}
	}

	// create array with the range of lines
	lines := make([]LineText, 0)
	for lineNumber := line1; lineNumber <= line2; lineNumber++ {
		lines = append(lines, LineText{Number: lineNumber})
	}

	return lines, nil
}
