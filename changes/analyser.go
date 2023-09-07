package changes

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
)

type ChangesOptions struct {
	utils.BaseOptions
	// AuthorsRegex string
	SinceDate   string
	UntilDate   string
	SinceCommit string
	UntilCommit string
}

type ChangesTimeseriesOptions struct {
	utils.BaseOptions
	Since  string `json:"since"`
	Until  string `json:"until"`
	Period string `json:"period"`
}

type LinesTouched struct {
	/* New lines found in commits */
	New int

	/* Lines changed in commits. If the same line is changed in two commits, for example, it will count as two changes. This is the sum of RefactorOwn, RefactorOther, ChurnOwn and ChurnOther */
	Changes int

	/* Lines changed after a while in which the author of the previous version was the same person */
	RefactorOwn int
	/* Lines changed after a while in which the author of the previous version was another person */
	RefactorOther int
	/* Lines you owned that were changed by another person after a while. When adding RefactorOther to someone, the author of the previous version of the line will have this counter incremented */
	RefactorReceived int

	/* Lines changed in a short term in which the author of the previous version was the same person */
	ChurnOwn int
	/* Lines changed in a short term in which the author of the previous version was another person */
	ChurnOther int
	/* Lines you owned that were changed by another person in a short term. When adding ChurnOther to someone, the author of the previous version of the line will have this counter incremented */
	ChurnReceived int

	/* Sum of age of lines in the moment they are changed. AgeDaysSum/Changes gives you the average survival duration of a line before it's changed by someone */
	AgeDaysSum float64
}

type FileTouched struct {
	Name  string
	Lines int
}

type AuthorLines struct {
	AuthorName      string
	AuthorMail      string
	LinesTouched    LinesTouched
	FilesTouched    []FileTouched
	filesTouchedMap map[string]FileTouched // temporary map used during processing
}

type ChangesFileResult struct {
	CommitId string
	FilePath string
	ChangesResult
}

type ChangesResult struct {
	/* Lines change stats */
	TotalLinesTouched LinesTouched
	/* Total files changed in the different commits. If the same file is changed in two commits, for example, it will count as one. */
	TotalFiles int
	/* Number of commits analysed */
	TotalCommits   int
	authorLinesMap map[string]AuthorLines // temporary map used during processing
	/* Change stats per author */
	AuthorsLines  []AuthorLines
	SinceCommit   utils.CommitInfo
	UntilCommit   utils.CommitInfo
	analysisTime  time.Duration
	skippedFiles  int
	authorSkipped bool
}

type fileWorkerRequest struct {
	repoDir         string
	commitId        string
	filePath        string
	authorsRegex    string
	authorsNotRegex string
}
type commitWorkerRequest struct {
	repoDir  string
	commitId string
}

func AnalyseTimeseriesChanges(opts ChangesTimeseriesOptions, progressChan chan<- utils.ProgressInfo) ([]ChangesResult, error) {
	if opts.Period == "" {
		return nil, fmt.Errorf("opts.Period is required")
	}
	if opts.Until == "" {
		return nil, fmt.Errorf("opts.Until is required")
	}

	result := make([]ChangesResult, 0)
	until := opts.Until
	since := fmt.Sprintf("%s - %s", until, opts.Period)
	analysisOpts := ChangesOptions{
		BaseOptions: opts.BaseOptions,
	}

	processedCommits := make([]string, 0)

	for {
		// FIND "SINCE" COMMIT
		// see if the outer "since" is outside inner "since"
		sinceCommits, err := utils.ExecGetCommitsInDateRange(opts.RepoDir, opts.Branch, opts.Since, since)
		if err != nil {
			return nil, err
		}

		// find next commit that wasn't processed yet
		// this is necessary because git does a "loose" lookup for commits when using relative time periods
		// and we don't want to repeat the same commit in multiple periods (to avoid double couting)
		var sinceCommit *utils.CommitInfo
		for _, scommit := range sinceCommits {
			if !slices.Contains(processedCommits, scommit.CommitId) {
				sinceCommit = &scommit
				break
			}
		}

		// no unprocessed commits for since anymore
		if sinceCommit == nil {
			break
		}

		// FIND "UNTIL" COMMIT
		// find next commit that wasn't processed yet
		// this is necessary because git does a "loose" lookup for commits when using relative time periods
		// and we don't want to repeat the same commit in multiple periods (to avoid double couting)
		untilCommits, err := utils.ExecGetCommitsInDateRange(opts.RepoDir, opts.Branch, sinceCommit.Date.Format(time.RFC3339), until)
		if err != nil {
			return nil, err
		}

		var untilCommit *utils.CommitInfo
		for _, ucommit := range untilCommits {
			if !slices.Contains(processedCommits, ucommit.CommitId) {
				untilCommit = &ucommit
				break
			}
		}

		// probably there is no data in the period, so skip this and try next period
		// git supports multiple relative date arguments, so append it
		// if untilCommit == nil ||
		// 	(sinceCommit.CommitId == untilCommit.CommitId) {
		if untilCommit == nil {
			since = fmt.Sprintf("%s - %s", since, opts.Period)
			until = fmt.Sprintf("%s - %s", until, opts.Period)
			logrus.Debugf("Skipping period without commits. since=%s; until=%s", since, until)
			continue
		}

		analysisOpts.SinceCommit = sinceCommit.CommitId
		analysisOpts.UntilCommit = untilCommit.CommitId
		onwershipResult, err := AnalyseChanges(analysisOpts, progressChan)
		if err != nil {
			return nil, err
		}
		result = append(result, onwershipResult)

		// add all commits in range as processed
		processedCommits = appendProcessed(processedCommits, sinceCommit.CommitId)
		rangeCommitIds, err := utils.ExecCommitIdsInCommitRange(opts.RepoDir, opts.Branch, sinceCommit.CommitId, untilCommit.CommitId)
		for _, rangeid := range rangeCommitIds {
			processedCommits = appendProcessed(processedCommits, rangeid)
		}

		// git doesn't support relative times mixed with date time, only with date
		until = fmt.Sprintf("%s", sinceCommit.Date.Format(time.DateOnly))
		since = fmt.Sprintf("%s - %s", until, opts.Period)
		logrus.Debugf("Analysing since=%s until=%s", since, until)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].SinceCommit.Date.Before(result[j].SinceCommit.Date)
	})

	return result, nil
}

func appendProcessed(processedCommits []string, newId string) []string {
	if !slices.Contains(processedCommits, newId) {
		processedCommits = append(processedCommits, newId)
	}
	return processedCommits
}

func AnalyseChanges(opts ChangesOptions, progressChan chan<- utils.ProgressInfo) (ChangesResult, error) {

	// check if cached results exists
	if opts.CacheFile != "" {
		cachedResults, err := GetFromCache(opts)
		if err != nil {
			return ChangesResult{}, err
		}
		if cachedResults != nil {
			return *cachedResults, nil
		}
	}

	result := ChangesResult{
		TotalLinesTouched: LinesTouched{},
		authorLinesMap:    make(map[string]AuthorLines, 0),
		AuthorsLines:      make([]AuthorLines, 0)}

	progressInfo := utils.ProgressInfo{}

	fileCounterMap := make(map[string]bool, 0)

	if opts.Branch == "" {
		return ChangesResult{}, fmt.Errorf("opts.Branch is required")
	}

	logrus.Debugf("Analysing changes in branch %s from %s%s to %s%s", opts.Branch, opts.SinceDate, opts.SinceCommit, opts.UntilDate, opts.UntilCommit)

	fre, err := regexp.Compile(opts.FilesRegex)
	if err != nil {
		return result, errors.New("files filter regex is invalid. err=" + err.Error())
	}

	freNot, err := regexp.Compile(opts.FilesNotRegex)
	if err != nil {
		return result, errors.New("files-not filter regex is invalid. err=" + err.Error())
	}

	nrWorkers := runtime.NumCPU() - 1
	// nrWorkers := 1

	// MAP REDUCE - analyse files in parallel goroutines
	// we need to start workers in the reverse order so that all the chain
	// is prepared when submitting tasks to avoid deadlocks
	logrus.Debugf("Preparing a pool of workers to process file analysis in parallel")
	fileWorkersInputChan := make(chan fileWorkerRequest, 5000)
	fileWorkersOutputChan := make(chan ChangesFileResult, 5000)
	fileWorkersErrChan := make(chan error, nrWorkers)

	commitWorkersInputChan := make(chan commitWorkerRequest, 5000)

	// REDUCE - summarise counters (STEP 4/4)
	var summaryWorkerWaitGroup sync.WaitGroup
	summaryWorkerWaitGroup.Add(1)
	go func() {
		defer summaryWorkerWaitGroup.Done()

		commitsWithFiles := make(map[string]bool, 0)

		logrus.Debugf("Counting total lines changed per author")
		for fileResult := range fileWorkersOutputChan {

			if !fileResult.authorSkipped {
				commitsWithFiles[fileResult.CommitId] = true
				_, ok := fileCounterMap[fileResult.FilePath]
				if !ok {
					fileCounterMap[fileResult.FilePath] = true
					result.TotalFiles++
				}
				result.TotalLinesTouched = SumLinesTouched(result.TotalLinesTouched, fileResult.TotalLinesTouched)
				for author := range fileResult.authorLinesMap {
					fileAuthorLines := fileResult.authorLinesMap[author]
					authorLines := result.authorLinesMap[author]
					authorLines.LinesTouched = SumLinesTouched(authorLines.LinesTouched, fileAuthorLines.LinesTouched)
					authorLines.filesTouchedMap = sumFilesTouched(authorLines.filesTouchedMap, fileAuthorLines.filesTouchedMap)
					result.authorLinesMap[author] = authorLines
				}
			}

			progressInfo.CompletedTotalTime += fileResult.analysisTime
			progressInfo.CompletedTasks += 1 + fileResult.skippedFiles
			progressInfo.CompletedTotalTime += result.analysisTime
			progressInfo.Message = fmt.Sprintf("%s", fileResult.FilePath)
			if progressChan != nil {
				progressChan <- progressInfo
			}
		}

		result.TotalCommits = len(commitsWithFiles)

		logrus.Debugf("Preparing summary for each author")
		authorsLines := make([]AuthorLines, 0)
		for authorKeys := range result.authorLinesMap {
			authorLines := result.authorLinesMap[authorKeys]
			authorParts := strings.Split(authorKeys, "###")

			filesTouched := make([]FileTouched, 0)
			for filesKey := range authorLines.filesTouchedMap {
				filesTouched = append(filesTouched, authorLines.filesTouchedMap[filesKey])
			}

			authorsLines = append(authorsLines, AuthorLines{
				AuthorName:   authorParts[0],
				AuthorMail:   authorParts[1],
				LinesTouched: authorLines.LinesTouched,
				FilesTouched: filesTouched,
			})
		}

		sort.Slice(authorsLines, func(i, j int) bool {
			ai := authorsLines[i].LinesTouched
			aj := authorsLines[j].LinesTouched
			return ai.New+ai.Changes > aj.New+aj.Changes
		})
		result.AuthorsLines = authorsLines

	}()

	// MAP - file analysis workers (STEP 3/4)
	var fileWorkersWaitGroup sync.WaitGroup
	for i := 0; i < nrWorkers; i++ {
		fileWorkersWaitGroup.Add(1)
		go fileAnalysisWorker(fileWorkersInputChan, fileWorkersOutputChan, fileWorkersErrChan, &fileWorkersWaitGroup)
	}
	logrus.Debugf("Launched %d workers for analysis", nrWorkers)

	// MAP - commits workers (STEP 2/4)
	// for each commit discover changed files and submit then for analysis
	totalFiles := 0
	progressInfo.TotalTasksKnown = false

	var commitWorkersWaitGroup sync.WaitGroup
	for i := 0; i < nrWorkers; i++ {
		commitWorkersWaitGroup.Add(1)

		go func() {
			defer commitWorkersWaitGroup.Done()
			for req := range commitWorkersInputChan {
				// logrus.Debugf("Analysing commit %s", req.commitId)
				files, err := utils.ExecDiffTree(req.repoDir, req.commitId)
				if err != nil {
					logrus.Errorf("Error getting files changed in commit. err=%s", err)
					panic(5)
				}

				for _, fileName := range files {
					if strings.Trim(fileName, " ") == "" || !fre.MatchString(fileName) || (opts.FilesNotRegex != "" && freNot.MatchString(fileName)) {
						// logrus.Debugf("Ignoring file %s", fileName)
						continue
					}
					totalFiles += 1
					progressInfo.TotalTasks += 1
					fileWorkersInputChan <- fileWorkerRequest{
						repoDir:         opts.RepoDir,
						filePath:        fileName,
						commitId:        req.commitId,
						authorsRegex:    opts.AuthorsRegex,
						authorsNotRegex: opts.AuthorsNotRegex,
					}
				}
			}
		}()
	}

	// MAP - submit commits for analysis (STEP 1/4)
	// for each commit between 'from' and 'to' date, submit to be analysed
	if (opts.SinceDate != "" || opts.UntilDate != "") && (opts.SinceCommit != "" || opts.UntilCommit != "") {
		return result, fmt.Errorf("Cannot mix opts.SinceDate/UntilDate with opts.SinceCommit/UntilCommit")
	}

	// find commit ids from dates
	if opts.SinceDate != "" || opts.UntilDate != "" {
		since := opts.SinceDate
		if opts.SinceDate == opts.UntilDate {
			// allow git query to find commit at "until"
			since = ""
		}
		commitIds, err := utils.ExecCommitIdsInDateRange(opts.RepoDir, opts.Branch, since, opts.UntilDate)
		if err != nil {
			return result, err
		}
		opts.SinceCommit = commitIds[len(commitIds)-1]
		opts.UntilCommit = commitIds[0]
	}

	since := opts.SinceCommit
	if opts.SinceCommit == opts.UntilCommit {
		// allow git query to find commit at "until"
		since = ""
	}
	commits, err := utils.ExecGetCommitsInCommitRange(opts.RepoDir, opts.Branch, since, opts.UntilCommit)
	if err != nil {
		return result, err
	}
	commitIds := utils.CommitInfoToCommitIds(commits)

	if since != "" {
		// add since commit id at the begining of the list
		commitIds = append([]string{since}, commitIds...)
	}

	if len(commitIds) == 0 {
		return result, fmt.Errorf("No changes found")
	}

	// commits are in reverse order
	sinceCommit, err := utils.ExecGitCommitInfo(opts.RepoDir, commitIds[len(commitIds)-1])
	if err != nil {
		return result, fmt.Errorf("Error getting since commit. err=%s", err)
	}
	result.SinceCommit = sinceCommit

	untilCommit, err := utils.ExecGitCommitInfo(opts.RepoDir, commitIds[0])
	if err != nil {
		return result, fmt.Errorf("Error getting until commit. err=%s", err)
	}
	result.UntilCommit = untilCommit

	logrus.Debug("Sending commits to workers")
	for _, commitId := range commitIds {
		commitWorkersInputChan <- commitWorkerRequest{
			repoDir:  opts.RepoDir,
			commitId: commitId,
		}
	}
	close(commitWorkersInputChan)
	logrus.Debug("Finished sending commits to workers")

	// COMMIT WORKERS FINISHED
	commitWorkersWaitGroup.Wait()
	logrus.Debug("Commit workers finished")
	close(fileWorkersInputChan)
	progressInfo.TotalTasksKnown = true
	if progressChan != nil {
		progressChan <- progressInfo
	}
	logrus.Debugf("%d files scheduled for analysis", totalFiles)

	// FILE WORKERS FINISHED
	fileWorkersWaitGroup.Wait()
	logrus.Debug("File workers finished")
	close(fileWorkersOutputChan)
	close(fileWorkersErrChan)

	for workerErr := range fileWorkersErrChan {
		logrus.Errorf("Error during analysis. err=%s", workerErr)
		panic(2)
	}

	summaryWorkerWaitGroup.Wait()
	logrus.Debug("Summary worker finished")

	if opts.CacheFile != "" {
		SaveToCache(opts, result)
	}

	return result, nil
}

func sumFilesTouched(map1 map[string]FileTouched, map2 map[string]FileTouched) map[string]FileTouched {
	if map1 == nil {
		map1 = make(map[string]FileTouched, 0)
	}
	for fileNameKey := range map2 {
		fileChanges1 := map1[fileNameKey]
		fileChanges2 := map2[fileNameKey]
		fileChanges1.Name = fileChanges2.Name
		fileChanges1.Lines += fileChanges2.Lines
		map1[fileNameKey] = fileChanges1
	}
	return map1
}
