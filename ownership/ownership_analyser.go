package ownership

type OwnershipOptions struct {
	Branch string
	When   string
	Files  string
}

type AuthorLines struct {
	Author     string
	OwnedLines uint
}

type OwnershipResult struct {
	TotalLines  uint
	AuthorLines []AuthorLines
}

func AnalyseCodeOwnership(opts OwnershipOptions) OwnershipResult {
	result := OwnershipResult{}
	return result
}
