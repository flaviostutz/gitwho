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

	skippedFiles := 0
	for req := range analyseFileInputChan {

		// fmt.Printf(">>>>%s %s\n", req.commitId, req.filePath)

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
			// can't get file size when the file was deleted by commit, so it's not present anymore
			// TODO get previous version of the file and count these lines as "changed" because they were deleted?
			skippedFiles++
			continue
		}
		if fsize > 80000 {
			logrus.Debugf("Ignoring file because it's too big. file=%s, size=%d", req.filePath, fsize)
			skippedFiles++
			continue
		}

		isBin, err := utils.ExecDiffIsBinary(req.repoDir, req.commitId, req.filePath)
		if err != nil {
			analyseFileErrChan <- errors.New(fmt.Sprintf("Couldn't determine if file is binary. file=%s; commitId=%s; err=%s", req.filePath, req.commitId, err))
			break
		}
		if isBin {
			logrus.Debugf("Ignoring binary file. file=%s, commitId=%s", req.filePath, req.commitId)
			skippedFiles++
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
			logrus.Infof("Couldn't git blame cur version of file. Ignoring it. file=%s; commitId=%s", req.filePath, req.commitId)
			skippedFiles++
			continue
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
					dstBlame.AuthorMail,
					LinesChanges{New: 1})
			}
			changesFileResult.analysisTime = time.Since(startTime)
			analyseFileOutputChan <- changesFileResult
			continue
		}

		// blame previous version of the file (so we can compare from->to contents)
		fileSrcBlame, err := utils.ExecGitBlame(req.repoDir, req.filePath, prevCommitId)
		if err != nil {
			// analyseFileErrChan <- errors.New(fmt.Sprintf("Error on git blame prev. file=%s. err=%s", req.filePath, err))
			// break
			logrus.Infof("Couldn't git blame prev version of file. Ignoring it. file=%s; commitId=%s", req.filePath, prevCommitId)
			skippedFiles++
			continue
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
					fileDstBlame[diff.DstLines[0].Number-1].AuthorMail,
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
				dstAuthorMail := commitInfo.AuthorMail
				if diff.Operation == utils.OperationChange {
					dstAuthorName = fileDstBlame[diff.DstLines[0].Number-1].AuthorName
					dstAuthorMail = fileDstBlame[diff.DstLines[0].Number-1].AuthorMail
				}

				// REFACTOR - changes to "old" lines
				if lineAge > 504*time.Hour {

					// REFACTORED it's own line
					if srcline.AuthorName == commitInfo.AuthorName {
						addAuthorLines(&changesFileResult,
							dstAuthorName,
							dstAuthorMail,
							LinesChanges{
								Changes:     1,
								RefactorOwn: 1,
								AgeDaysSum:  commitInfo.Date.Sub(srcline.AuthorDate).Hours() / float64(24),
							})
						continue
					}

					// REFACTORED someone else's line
					addAuthorLines(&changesFileResult,
						dstAuthorName,
						dstAuthorMail,
						LinesChanges{
							Changes:       1,
							RefactorOther: 1,
							AgeDaysSum:    commitInfo.Date.Sub(srcline.AuthorDate).Hours() / float64(24),
						})

					// if someone changed your line, you receive a "refactor received" count
					addAuthorLines(&changesFileResult,
						srcline.AuthorName,
						srcline.AuthorMail,
						LinesChanges{RefactorReceived: 1})

					continue
				}

				// CHURN - changes to "young" lines

				// churn by the same author
				if srcline.AuthorName == commitInfo.AuthorName {
					addAuthorLines(&changesFileResult,
						dstAuthorName,
						dstAuthorMail,
						LinesChanges{
							Changes:    1,
							ChurnOwn:   1,
							AgeDaysSum: commitInfo.Date.Sub(srcline.AuthorDate).Hours() / float64(24),
						})
					continue
				}

				// churn by a different author
				addAuthorLines(&changesFileResult,
					dstAuthorName,
					dstAuthorMail,
					LinesChanges{
						Changes:    1,
						ChurnOther: 1,
						AgeDaysSum: commitInfo.Date.Sub(srcline.AuthorDate).Hours() / float64(24),
					})

				// if someone changed your line, you receive a "churn received" count
				addAuthorLines(&changesFileResult,
					srcline.AuthorName,
					srcline.AuthorMail,
					LinesChanges{ChurnReceived: 1})
			}

			// special case when changes led to additional lines in destination
			if len(diff.DstLines) > len(diff.SrcLines) {
				for i := 0; i < len(diff.SrcLines); i++ {
					dstline := fileDstBlame[i+diff.DstLines[0].Number-1]
					addAuthorLines(&changesFileResult,
						dstline.AuthorName,
						dstline.AuthorMail,
						LinesChanges{New: 1})
				}
			}

		}

		changesFileResult.analysisTime = time.Since(startTime)
		changesFileResult.skippedFiles = skippedFiles
		skippedFiles = 0
		analyseFileOutputChan <- changesFileResult
		// time.Sleep(1 * time.Second)
		// fmt.Printf("Time spent: %s\n", time.Since(startTime))
	}
}

func addAuthorLines(changesFileResult *ChangesFileResult, authorName string, authorMail string, linesChanges LinesChanges) {
	authorKey := fmt.Sprintf("%s###%s", authorName, authorMail)
	authorLine := changesFileResult.authorLinesMap[authorKey]
	// add to author totals
	authorLine = sumLinesChanges(authorLine, linesChanges)
	changesFileResult.authorLinesMap[authorKey] = authorLine

	// add to overall totals
	changesFileResult.TotalLines = sumLinesChanges(changesFileResult.TotalLines, linesChanges)
}
