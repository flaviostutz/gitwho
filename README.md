# gitwho

Shows statistics about changed lines and line ownership per authors in a git repo. Take a look at the team or individual devs to understand better their behaviour and their evolution over time.

This utility gives you stats like:
  - Changed lines per author over time classified in new, churn, helper or refactor (see concepts below)
  - Files with most new, churn, helper or refactor changes
  - Total owned lines of code per author in a moment in time

## Usage

* Download the executable binary on releases page and execute `chmod +x gitwho`

```sh
cd myrepo

# show changes stats for the last 30 days
npx gitwho changes

# show ownership stats for today
npx gitwho ownership

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

