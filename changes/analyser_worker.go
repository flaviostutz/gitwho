package changes

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/flaviostutz/gitwho/utils"
	"github.com/sirupsen/logrus"
)

// this will be run by multiple goroutines
func analyseFileChangesWorker(analyseFileInputChan <-chan analyseFileRequest, analyseFileOutputChan chan<- ChangesFileResult, analyseFileErrChan chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	for req := range analyseFileInputChan {

		fmt.Printf(">>>>%s %s\n", req.commitId, req.filePath)

		startTime := time.Now()
		changesFileResult := ChangesFileResult{
			CommitId: req.commitId,
			FilePath: req.filePath,
			ChangesResult: ChangesResult{
				TotalLines:     LinesChanges{},
				authorLinesMap: make(map[string]LinesChanges, 0),
				AuthorsLines:   []AuthorLines{},
			},
		}

		fsize, err := utils.ExecTreeFileSize(req.repoDir, req.commitId, req.filePath)
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Couldn't find file. file=%s. err=%s", req.filePath, err))
			break
		}
		if fsize > 80000 {
			logrus.Debugf("Ignoring file because it's too big. file=%s, size=%d", req.filePath, fsize)
			continue
		}

		commitInfo, err := utils.ExecGitCommitInfo(req.repoDir, req.commitId)
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Couldn't get commit info. commitId=%s; err=%s", req.commitId, err))
			break
		}

		// blame current version of the file
		fileDstBlame, err := utils.ExecGitBlame(req.repoDir, req.filePath, req.commitId)
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Error on git blame cur. file=%s. err=%s", req.filePath, err))
			break
		}

		// find the previous commit in which this file was changed
		prevCommitId, err := utils.ExecPreviousCommitIdForFile(req.repoDir, req.commitId, req.filePath)
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Error on getting prev commit id. err=%s", err))
			break
		}

		changesFileResult.TotalFiles += 1

		// there is no previous commit because this is a brand new file
		if prevCommitId == "" {
			// consider all lines as "New"
			for _, dstBlame := range fileDstBlame {
				addAuthorLines(&changesFileResult,
					dstBlame.AuthorName,
					LinesChanges{New: 1})
			}
			changesFileResult.analysisTime = time.Since(startTime)
			analyseFileOutputChan <- changesFileResult
			continue
		}

		// blame previous version of the file (so we can compare from->to contents)
		fileSrcBlame, err := utils.ExecGitBlame(req.repoDir, req.filePath, prevCommitId)
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Error on git blame prev. file=%s. err=%s", req.filePath, err))
			break
		}

		// diff both versions of the file
		// diffs := diffMatcher.DiffMain(filePrevContents, fileCurContents, false)
		diffs, err := utils.ExecDiffFileRevisions(req.repoDir, req.filePath, prevCommitId, req.commitId)
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Couldn't diff file revisions. file=%s; srcCommit=%s; dstCommit=%s; err=%s", req.filePath, prevCommitId, req.commitId, err))
			break
		}

		// for each line, classify change type
		for _, diff := range diffs {

			// NEW lines
			if diff.Operation == utils.OperationAdd {
				// added lines are simply "new"
				addAuthorLines(&changesFileResult,
					fileDstBlame[diff.DstLines[0].Number-1].AuthorName,
					LinesChanges{New: len(diff.DstLines)})
				continue
			}

			// CHANGED lines
			// classify change as
			//   REFACTOR - when the line changed was more than 21 days old
			//   CHURN - when the line changed was less than 21 days old
			for i := 0; i < len(diff.SrcLines); i++ {
				srcline := fileSrcBlame[i+diff.SrcLines[0].Number-1]
				lineAge := commitInfo.Date.Sub(srcline.AuthorDate)

				dstAuthorName := commitInfo.AuthorName
				if diff.Operation == utils.OperationChange {
					dstAuthorName = fileDstBlame[diff.DstLines[0].Number-1].AuthorName
				}

				// REFACTOR - changes to "old" lines
				if lineAge > 504*time.Hour {

					// REFACTORED it's own line
					if srcline.AuthorName == commitInfo.AuthorName {
						addAuthorLines(&changesFileResult,
							dstAuthorName,
							LinesChanges{
								RefactorOwn: 1,
								AgeSum:      commitInfo.Date.Sub(srcline.AuthorDate),
							})
						continue
					}

					// REFACTORED someone else's line
					addAuthorLines(&changesFileResult,
						dstAuthorName,
						LinesChanges{
							Changes:       1,
							RefactorOther: 1,
							AgeSum:        commitInfo.Date.Sub(srcline.AuthorDate),
						})

					// if someone changed your line, you receive a "refactor received" count
					addAuthorLines(&changesFileResult,
						srcline.AuthorName,
						LinesChanges{RefactorReceived: 1})

					continue
				}

				// CHURN - changes to "young" lines

				// churn by the same author
				if srcline.AuthorName == commitInfo.AuthorName {
					addAuthorLines(&changesFileResult,
						dstAuthorName,
						LinesChanges{
							Changes:  1,
							ChurnOwn: 1,
							AgeSum:   commitInfo.Date.Sub(srcline.AuthorDate),
						})
					continue
				}

				// churn by a different author
				addAuthorLines(&changesFileResult,
					dstAuthorName,
					LinesChanges{
						Changes:    1,
						ChurnOther: 1,
						AgeSum:     commitInfo.Date.Sub(srcline.AuthorDate),
					})

				// if someone changed your line, you receive a "churn received" count
				addAuthorLines(&changesFileResult,
					srcline.AuthorName,
					LinesChanges{RefactorReceived: 1})
			}

			// special case when changes led to additional lines in destination
			if len(diff.DstLines) > len(diff.SrcLines) {
				for i := 0; i < len(diff.SrcLines); i++ {
					dstline := fileDstBlame[i+diff.DstLines[0].Number-1]
					addAuthorLines(&changesFileResult,
						dstline.AuthorName,
						LinesChanges{New: 1})
				}
			}

		}

		changesFileResult.analysisTime = time.Since(startTime)
		analyseFileOutputChan <- changesFileResult
		// time.Sleep(1 * time.Second)
		// fmt.Printf("Time spent: %s\n", time.Since(startTime))
	}
}

func addAuthorLines(changesFileResult *ChangesFileResult, authorName string, linesChanges LinesChanges) {
	authorLine := changesFileResult.authorLinesMap[authorName]
	// add to author totals
	authorLine = sumLinesChanges(authorLine, linesChanges)
	changesFileResult.authorLinesMap[authorName] = authorLine

	// addo to overall totals
	changesFileResult.TotalLines = sumLinesChanges(changesFileResult.TotalLines, linesChanges)
}
