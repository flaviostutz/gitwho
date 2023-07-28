# gitwho

Enables you to check the overal classification of the changes on the codebase, or to take a look at individual devs to understand better their behaviour and their evolution over time.

This utility gives you stats like:
  - Changed lines per author over time classified in new, churn, helper or refactor (see concepts below)
  - Files with most new, churn, helper or refactor changes
  - Total owned lines of code per author in a moment in time

## Usage

* Download the executable binary on releases page and execute `chmod +x gitwho`

```sh
cd myrepo

# show stats for the last 90 days
gitwho --days 90

```

### gitwho ownership

* Gets the current situation of a repository in a moment in time and counts how many lines of code was created by whom by doing git blame in all files in the repo. For more info, check https://git-scm.com/docs/git-blame


## gitwho changes

* Go through all the commits in a certain period and classify which kind of change was done to the lines changed. The final result is the sum of all changes, so for example, if the same line was touched in 4 commits, it will show as 4 lines changed in total. The idea is to show the running effort during coding.

## Types of change concept

When a line is added or deleted by a commit, the context of the change will be analysed so we can classify it in:

- *Code churn*
  - Lines that were changed by the same dev multiple times in a very short period
  - Might indicate code that is unstable, buggy or that gets pushed to repo too soon, before minimum quality

- *Code help*
  - Lines that were changed by a different dev multiple times in a very short period
  - Might indicate someone helping other devs to strive/meeting their goals by removing bugs, enhancing code from newly created features etc

- *Code refactoring*
  - Lines that were changed after a while
  - Might indicate code that was improved because of tech debt resolutions, continuous improvement behaviours or hard to tackle bugs found after the code was being used for a while

- *New Code*
  - New lines of code
  - Might indicate new features being added or spikes being made

See more info in this excelent article: https://www.hatica.io/blog/code-churn-rate/

## Analysis details

- Blank lines are excluded from all analysis

## Development notes

// get number of lines last created per author in total with blame
git ls-tree --name-only -z -r HEAD|egrep -z -Z -E '\.(cc|h|cpp|hpp|c|txt|ts|js)$' \
  |xargs -0 -n1 git blame --line-porcelain|grep "^author "|sort|uniq -c|sort -nr


// go-git checked out dir
https://github.com/go-git/go-git/blob/master/_examples/open/main.go

https://github.com/sergi/go-diff

// list commit ids, date, author
git log --pretty=format:"%H %cI %an"

// show commit logs
git log --reverse --numstat



Remove from analysis known auto-generated files? (yarn.lock go.sum etc)


https://softwareengineering.stackexchange.com/questions/429319/legacy-refactor-and-churn
https://github.com/andymeneely/git-churn



In a specific moment, per author:
  - number of lines owned

Between two moments, per author:

  - number of lines with new code
     - line doesn’t exist in the previous version

  - number of lines with refactored code
     - the previous version of the line is more than 21 days old and was changed

  - number of lines with helping code
     - the previous version of the line is less than 21 days old from another author and was changed or deleted by the author

  - number of lines with churn code
     - the previous version of the line is less than 21 days old from the same author and it was changed or deleted


// show all commits that touched a certain file
git log --all package.json

// show blame at revision point
git blame 3a974275d -- package.json

// show blame at revision parent point
git blame 3a974275d^ -- package.json

// show dir tree at a revision point
git ls-tree 3a974275d

