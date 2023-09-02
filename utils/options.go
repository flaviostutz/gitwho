package utils

type BaseOptions struct {
	Branch          string `json:"branch"`
	FilesRegex      string `json:"files_regex"`
	FilesNotRegex   string `json:"files_not_regex"`
	AuthorsRegex    string `json:"authors_regex"`
	AuthorsNotRegex string `json:"authors_not_regex"`
	RepoDir         string `json:"repo_dir"`
	CacheFile       string `json:"cache_file"`
}
