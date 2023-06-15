package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/package-url/packageurl-go"

	"github.com/ossf/package-analysis/internal/worker"
)

// Command-line tool to download a list of package archives, specified by purl
// See https://github.com/package-url/purl-spec
var (
	purlFilePath = flag.String("f", "", "file containing packages to download")
	downloadDir  = flag.String("d", "", "directory to download files to (must exist)")
)

// usageError is a simple string error type, used when command usage
// should be printed alongside the actual error message
type cmdError struct {
	message string
}

func (c *cmdError) Error() string {
	return c.message
}

func newCmdError(message string) error {
	return &cmdError{message}
}

func downloadPackage(purl packageurl.PackageURL, downloadDir string) error {
	pkg, err := worker.ResolvePurl(purl)

	if err != nil {
		return err
	}

	fmt.Printf("[%s] %s@%s\n", pkg.EcosystemName(), pkg.Name(), pkg.Version())

	if _, err := pkg.Manager().DownloadArchive(pkg.Name(), pkg.Version(), downloadDir); err != nil {
		return err
	}

	return nil
}

func checkDirectoryExists(path string) error {
	stat, err := os.Stat(path)

	if err != nil && errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("path %s does not exist", path)
	}
	if err != nil {
		return fmt.Errorf("could not stat %s: %w", path, err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", path)
	}

	return nil
}

func processFileLine(line int, text string) {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) == 0 || trimmed[0] == '#' {
		return
	}

	if purl, err := packageurl.FromString(trimmed); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing purl '%s' on line %d: %v", text, line, err)
	} else if err := downloadPackage(purl, *downloadDir); err != nil {
		fmt.Fprintf(os.Stderr, "error downloading '%s': %v", text, err)
	}
}

func run() error {
	flag.Parse()

	if *purlFilePath == "" {
		return newCmdError("package list file (-f) is a required argument")
	}
	if *downloadDir == "" {
		*downloadDir = "."
	}

	if err := checkDirectoryExists(*downloadDir); err != nil {
		return err
	}

	purlFile, err := os.Open(*purlFilePath)
	if err != nil {
		return err
	}

	defer func() { _ = purlFile.Close() }()

	scanner := bufio.NewScanner(purlFile)
	line := 0
	for scanner.Scan() {
		line += 1
		processFileLine(line, scanner.Text())
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		var cmdErr *cmdError
		if errors.As(err, &cmdErr) {
			flag.Usage()
			fmt.Fprintf(os.Stderr, "\n")
		}
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
