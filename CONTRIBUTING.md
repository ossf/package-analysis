# Contributing to Package Analysis

Hello new contributor! Thank you for contributing your time and expertise to the Package Analysis project.
We're delighted to have you on board.

This document describes the contribution guidelines for the project.

## Ways to get in touch
If you have any contribution-related questions, please get in touch! Here are some ways to reach current contributors
1. Open a new issue (strongly preferred)
1. Via the [OpenSSF Securing Critical Projects Working Group](https://github.com/ossf/wg-securing-critical-projects) mailing list or Slack channel

Note: for minor changes (typos, documentation improvements), feel free to open a pull request directly.


**Note:** Before you start contributing, you must read and abide by our
**[Code of Conduct](./CODE_OF_CONDUCT.md)**.

## Contributing code

### Getting started

1.  Create [a GitHub account](https://github.com/join)
1.  Create a [personal access token](https://docs.github.com/en/free-pro-team@latest/developers/apps/about-apps#personal-access-tokens)
1.  (Optionally) a Google Cloud Platform account for [deps.dev](https://deps.dev) data
1.  Set up your [development environment](#environment-setup)

## Environment Setup

You must install these tools:

1.  [`git`](https://help.github.com/articles/set-up-git/): For source control.

1.  [`go`](https://go.dev/dl/): For running code.

And optionally:

1.  [`gcloud`](https://cloud.google.com/sdk/docs/install): For Google Cloud Platform access for deps.dev data.

Then clone the repository, e.g:

```shell
$ git clone git@github.com:ossf/package-analysis.git
$ cd package-analysis
```

## Notes on style
### Commit style:
Prefer small but frequent PRs to make reviewing easier.

If making many incremental changes, one way to avoid being blocked by reviews while still writing new code is to make a new branch off the branch under review, and opening a new PR which is diffed to the first branch. This can be chained multiple times. However, the commits of earlier PRs are duplicated in later PRs; this causes the default squashed commit message of the later PRs to contain commit messages from the earlier PRs. Rebasing branches of later PRs onto main after earlier PRs have been merged may solve this.

### Code style:

#### Warnings:
Some things that are OK:
- not handling the error when `defer` close() on an HTTP response body

#### Comments:
Follow official Go comment style: https://tip.golang.org/doc/comment.
In particular, all exported (capitalised) types and functions should have a comment explaining what they do.
The comment should start with the type/function name.


#### Imports:
- stdlib imports grouped first, then 3rd party packages, then local imports
- each group separated by a blank line and ordered alphabetically

##### on IntelliJ:
- Remove redundant import aliases: yes
- Sorting type: gofmt
- Move all imports into a single declaration: yes
- Group stdlib imports: yes
- Move all stdlib imports in a single group: yes
- Group: yes, current project packages

