package changes

import (
	"sync"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type ChangesOptions struct {
	utils.BaseOptions
	AuthorsRegex string
	Since        string
	Until        string
}

type LinesChanges struct {
	New      int
	Refactor int
	Churn    int
	Helper   int
}

type AuthorLines struct {
	Author string
	Lines  LinesChanges
}

type ChangesFileResult struct {
	ChangesResult
	FilePath  string
	CommitId  string
	blameTime time.Duration
}

type ChangesResult struct {
	TotalLines     LinesChanges
	TotalFiles     int
	authorLinesMap map[string]LinesChanges // temporary map used during processing
	AuthorsLines   []AuthorLines
}

type analyseFileRequest struct {
	filePath    string
	workingTree *git.Worktree
	fromCommit  *object.Commit
	toCommit    *object.Commit
}

func AnalyseChanges(repo *git.Repository, opts ChangesOptions, progressChan chan<- utils.ProgressInfo) (ChangesResult, error) {
	result := ChangesResult{
		TotalLines:     LinesChanges{},
		authorLinesMap: make(map[string]LinesChanges, 0),
		AuthorsLines:   make([]AuthorLines, 0)}

	// progressInfo := utils.ProgressInfo{}

	// fre, err := regexp.Compile(opts.FilesRegex)
	// if err != nil {
	// 	return result, errors.New("file filter regex is invalid. err=" + err.Error())
	// }

	// logrus.Debugf("Analysing changes in branch %s from %s to %s", opts.Branch, opts.From, opts.To)

	// // MAP REDUCE - analyse files in parallel goroutines
	// // we need to start workers in the reverse order so that all the chain
	// // is prepared when submitting tasks to avoid deadlocks
	// nrWorkers := runtime.NumCPU() - 1
	// logrus.Debugf("Preparing a pool of workers to process file analysis in parallel")
	// analyseFileInputChan := make(chan analyseFileRequest, 5000)
	// analyseFileOutputChan := make(chan ChangesFileResult, 5000)
	// analyseFileErrChan := make(chan error, nrWorkers)

	// // REDUCE - summarise counters (STEP 3/3)
	// var summaryWorkerWaitGroup sync.WaitGroup
	// summaryWorkerWaitGroup.Add(1)
	// go func() {
	// 	defer summaryWorkerWaitGroup.Done()
	// 	logrus.Debugf("Counting total lines changed per author")
	// 	for fileResult := range analyseFileOutputChan {
	// 		result.TotalFiles += fileResult.TotalFiles
	// 		result.TotalLines.Churn += fileResult.TotalLines.Churn
	// 		result.TotalLines.Helper += fileResult.TotalLines.Helper
	// 		result.TotalLines.New += fileResult.TotalLines.New
	// 		result.TotalLines.Refactor += fileResult.TotalLines.Refactor
	// 		for author := range fileResult.authorLinesMap {
	// 			fileAuthorLines := fileResult.authorLinesMap[author]
	// 			resultAuthorLines := result.authorLinesMap[author]
	// 			resultAuthorLines.Churn += fileAuthorLines.Churn
	// 			resultAuthorLines.Helper += fileAuthorLines.Helper
	// 			resultAuthorLines.New += fileAuthorLines.New
	// 			resultAuthorLines.Refactor += fileAuthorLines.Refactor
	// 		}
	// 		progressInfo.CompletedTasks += 1
	// 		progressInfo.Message = fmt.Sprintf("%s (%s)", fileResult.FilePath, fileResult.blameTime)
	// 		if len(progressChan) < 1 {
	// 			progressChan <- progressInfo
	// 		}
	// 	}

	// 	logrus.Debugf("Sorting and preparing summary for each author")
	// 	authorsLines := make([]AuthorLines, 0)
	// 	for author := range result.authorLinesMap {
	// 		lines := result.authorLinesMap[author]
	// 		authorsLines = append(authorsLines, AuthorLines{Author: author, Lines: lines})
	// 	}

	// 	sort.Slice(authorsLines, func(i, j int) bool {
	// 		ai := authorsLines[i].Lines
	// 		aj := authorsLines[j].Lines
	// 		return ai.Churn+ai.Helper+ai.New+ai.Refactor > aj.Churn+aj.Helper+aj.New+aj.Refactor
	// 	})
	// 	result.AuthorsLines = authorsLines
	// }()

	// // MAP - start analyser workers (STEP 2/3)
	// var analysisWorkersWaitGroup sync.WaitGroup
	// for i := 0; i < nrWorkers; i++ {
	// 	analysisWorkersWaitGroup.Add(1)
	// 	go analyseFileChangesWorker(analyseFileInputChan, analyseFileOutputChan, analyseFileErrChan, &analysisWorkersWaitGroup)
	// }
	// logrus.Debugf("Launched %d workers for analysis", nrWorkers)

	// // MAP - submit tasks (STEP 1/3)
	// go func() {

	// 	// for each commit between 'from' and 'to' date, discover changed files and submit then for analysis

	// 	// FOR ALL COMMITS BETWEEN
	// 	//    CHECKOUT WORKTREE
	// 	//    DISCOVER CHANGED FILES FROM PREVIOUS COMMIT
	// 	//    SUBMIT FILE FOR ANALYSIS

	// 	branchHash, err := utils.GetBranchHash(repo, opts.Branch)
	// 	if err != nil {
	// 		logrus.Errorf("Couldn't find branch %s. err=%s", opts.Branch, err)
	// 		// FIXME create a better error handling mechanism
	// 		panic(1)
	// 	}

	// 	commits, err := repo.Log(&git.LogOptions{Since: &opts.From, Until: &opts.To, From: branchHash})
	// 	if err != nil {
	// 		logrus.Error("Couldn't load commit list. err=%s", err)
	// 		// FIXME create a better error handling mechanism
	// 		panic(1)
	// 	}

	// 	// for each commit schedule file change analysis
	// 	err = commits.ForEach(func(c *object.Commit) error {
	// 		// wt, err := repo.Worktree()
	// 		// if err != nil {
	// 		// 	return err
	// 		// }
	// 		// logrus.Debugf("Checking out commit %s", c.Hash)
	// 		// err = wt.Checkout(&git.CheckoutOptions{Hash: c.Hash})
	// 		// if err != nil {
	// 		// 	return err
	// 		// }

	// 		// parei aqui... use isso pra identificar changed files
	// 		// currentTree, err := commit.Tree()
	// 		// CheckIfError(err)

	// 		// prevTree, err := prevCommit.Tree()
	// 		// CheckIfError(err)

	// 		// patch, err := currentTree.Patch(prevTree)
	// 		// CheckIfError(err)
	// 		// fmt.Println("----- Patch Stats ------")

	// 		// var changedFiles []string
	// 		// for _, fileStat := range patch.Stats() {
	// 		// 	fmt.Println(fileStat.Name)
	// 		// 	changedFiles = append(changedFiles,fileStat.Name)
	// 		// }

	// 		// changes, err := currentTree.Diff(prevTree)
	// 		// CheckIfError(err)
	// 		// fmt.Println("----- Changes -----")
	// 		// for _, change := range changes {
	// 		// 	// Ignore deleted files
	// 		// 	action, err := change.Action()
	// 		// 	CheckIfError(err)
	// 		// 	if action == merkletrie.Delete {
	// 		// 		//fmt.Println("Skipping delete")
	// 		// 		continue
	// 		// 	}
	// 		// 	// Get list of involved files
	// 		// 	name := getChangeName(change)
	// 		// 	fmt.Println(name)
	// 		// }

	// 		logrus.Debugf("Scheduling commit %s files for analysis. filesRegex=%s", c.Hash, opts.FilesRegex)
	// 		totalFiles := 0
	// 		progressInfo.TotalTasksKnown = false
	// 		fsutil.Walk(wt.Filesystem, "/", func(path string, finfo fs.FileInfo, err error) error {
	// 			// fmt.Printf("%s, %s, %s\n", path, finfo, err)
	// 			if finfo == nil || finfo.IsDir() || finfo.Size() > 30000 || !fre.MatchString(path) || strings.Contains(path, "/.git/") {
	// 				// logrus.Debugf("Ignoring file %s", finfo)
	// 				return nil
	// 			}
	// 			totalFiles += 1

	// 			// show progress
	// 			progressInfo.TotalTasks += 1
	// 			// if len(progressChan) < 1 {
	// 			// 	progressChan <- progressInfo
	// 			// }

	// 			// schedule file to be blamed by parallel workers
	// 			analyseFileInputChan <- analyseFileRequest{filePath: path, workingTree: wt, commitObj: commitObj}
	// 			return nil
	// 		})
	// 		return nil
	// 	})
	// 	if err != nil {
	// 		logrus.Errorf("Couldn't iterate commit list. err=%s", err)
	// 		// FIXME create a better error handling mechanism
	// 		panic(1)
	// 	}

	// 	// finished publishing request messages
	// 	logrus.Debugf("%d files scheduled for analysis", totalFiles)
	// 	logrus.Debug("Task submission worker finished")
	// 	close(analyseFileInputChan)

	// 	progressInfo.TotalTasksKnown = true
	// 	if len(progressChan) < 1 {
	// 		progressChan <- progressInfo
	// 	}
	// }()

	// analysisWorkersWaitGroup.Wait()
	// logrus.Debug("Analysis workers finished")
	// close(analyseFileOutputChan)
	// close(analyseFileErrChan)

	// for workerErr := range analyseFileErrChan {
	// 	logrus.Errorf("Error during analysis. err=%s", workerErr)
	// 	panic(2)
	// }

	// summaryWorkerWaitGroup.Wait()
	// logrus.Debug("Summary worker finished")

	// // fmt.Printf("SUMMARY: %v\n", result)

	return result, nil
}

// this will be run by multiple goroutines
func analyseFileChangesWorker(analyseFileInputChan <-chan analyseFileRequest, analyseFileOutputChan chan<- ChangesFileResult, analyseFileErrChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	// for req := range analyseFileInputChan {
	// 	startTime := time.Now()
	// 	ownershipResult := OwnershipResult{TotalLines: 0, authorLinesMap: make(map[string]int, 0)}
	// 	ownershipResult.FilePath = req.filePath
	// 	blameResult, err := git.Blame(req.commitObj, strings.TrimLeft(req.filePath, "/"))
	// 	if err != nil {
	// 		analyseFileErrChan <- errors.New(fmt.Sprintf("Error on git blame. file=%s. err=%s", req.filePath, err))
	// 		break
	// 	}
	// 	//TODO: IMPLEMENT IN GIT TO COMPARE SPEED
	// 	ownershipResult.TotalFiles += 1
	// 	for _, lineAuthor := range blameResult.Lines {
	// 		if strings.Trim(lineAuthor.Text, " ") == "" {
	// 			continue
	// 		}
	// 		ownershipResult.TotalLines += 1
	// 		ownershipResult.authorLinesMap[lineAuthor.AuthorName] += 1
	// 	}
	// 	ownershipResult.blameTime = time.Since(startTime)
	// 	analyseFileOutputChan <- ownershipResult
	// 	// time.Sleep(1 * time.Second)
	// 	// fmt.Printf("Time spent: %s\n", time.Since(startTime))
	// }
}
