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
1.  Set up your [development environment](#environment-setup)

## Environment Setup

You must install these tools:

1.  [`git`](https://help.github.com/articles/set-up-git/): For source control.
1.  [`go`](https://go.dev/dl/): For running code.
1.  `make`: For running development commands

For running/testing locally, the following additional tools are required:

1.  [`docker`](https://www.docker.com/get-started/): The external container
1.  [`podman`](https://podman.io/getting-started/): The internal container
1.  [`docker-compose`](https://docs.docker.com/compose/install/) for end-to-end testing

Then clone the repository, e.g:

```shell
$ git clone git@github.com:ossf/package-analysis.git
$ cd package-analysis
```

## Notes on style

### Commit style

Prefer smaller PRs to make reviewing easier. Larger changes can be split into smaller PRs by branching off previous (unmerged) branches rather than main.

### Code style

We generally follow the [Google Go Style Guide](https://google.github.io/styleguide/go/index).

#### Warnings

Some things that are OK:

- not handling the error when `defer` close() on an HTTP response body

#### Comments

Follow official Go comment style: https://tip.golang.org/doc/comment.
In particular, all exported (capitalised) types and functions should have a comment explaining what they do.
The comment should start with the type/function name.

#### Imports

- stdlib imports grouped first, then 3rd party packages, then local imports
- each group separated by a blank line and ordered alphabetically

##### on IntelliJ

- Remove redundant import aliases: yes
- Sorting type: gofmt
- Move all imports into a single declaration: yes
- Group stdlib imports: yes
- Move all stdlib imports in a single group: yes
- Group: yes, current project packages

