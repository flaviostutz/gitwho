package ownership

import "fmt"

func FormatTextResults(ownershipResult OwnershipResult, opts OwnershipOptions) string {
	text := "Code ownership analysis\n"
	text += fmt.Sprintf("Branch: %s at %s (%s)\n", opts.Branch, opts.WhenStr, ownershipResult.CommitId)
	filesStr := "all"
	if opts.FilesRegex != "" {
		filesStr = opts.FilesRegex
	}
	text += fmt.Sprintf("Files: %s\n", filesStr)
	text += "\n"
	text += fmt.Sprintf("Total lines of code: %d\n", ownershipResult.TotalLines)

	for _, authorLines := range ownershipResult.AuthorsLines {
		text += fmt.Sprintf("%s: %d\n", authorLines.Author, authorLines.OwnedLines)
	}
	return text
}
