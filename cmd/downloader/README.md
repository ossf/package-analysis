
# Package Download tool

This tool enables easy batch download of many packages to a local directory,
which may be useful for testing or running analysis locally.

## Building

```bash
go build -o downloader main.go
```

## Running

```bash
./downloader -f <packages.txt> -d <dir>
```

There are two options to the downloader tool:

1. List of packages to download (mandatory)
2. Destination directory to download to (optional)

If `-d` is not specified, packages will be downloaded to the current directory.

The file containing the list of packages to download must have the following structure:

1. Each line of the file specifies one package to download in
   [Package URL](https://github.com/package-url/purl-spec) format
2. Package ecosystem and name are required, version is optional
3. If the version is not given, the latest version is downloaded

Here are some examples of Package URLs (purls):

- `pkg:npm/async`: NPM package `async`, no version specified
- `pkg:pypi/requests@2.31.0`: PyPI package `requests`, version 2.31.0
- `pkg:npm/%40babel/runtime`: NPM package `@babel/runtime` (note: percent encoding is not required by this tool)

If Package URL is invalid or a package fails to download, the error will be printed but will not stop the program;
remaining package downloads will still be attempted.