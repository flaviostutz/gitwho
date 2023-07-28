package changes

func FormatTextResults(ownershipResult ChangesResult, opts ChangesOptions) string {
	return "TO BE DONE"
	// text := "Line ownership analysis\n"
	// text += "\n"
	// text += fmt.Sprintf("Branch: %s at %s (%s)\n", opts.Branch, opts.WhenStr, ownershipResult.CommitId)
	// filesStr := "all"
	// if opts.FilesRegex != "" {
	// 	filesStr = opts.FilesRegex
	// }
	// text += fmt.Sprintf("Files: %s\n", filesStr)
	// text += "\n"

	// text += fmt.Sprintf("Total authors: %d\n", len(ownershipResult.AuthorsLines))
	// text += fmt.Sprintf("Total files: %d\n", ownershipResult.TotalFiles)
	// text += fmt.Sprintf("Total lines: %d\n", ownershipResult.TotalLines)

	// text += "\n"
	// for _, authorLines := range ownershipResult.AuthorsLines {
	// 	text += fmt.Sprintf("%s: %d (%s%%)\n", authorLines.Author, authorLines.OwnedLines, strconv.FormatFloat(float64(100)*(float64(authorLines.OwnedLines)/float64(ownershipResult.TotalLines)), 'f', 1, 32))
	// }
	// return text
}
