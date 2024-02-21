package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/package-url/packageurl-go"

	"github.com/ossf/package-analysis/internal/useragent"
	"github.com/ossf/package-analysis/internal/worker"
)

// Command-line tool to download a list of package archives, specified by purl
// See https://github.com/package-url/purl-spec
var (
	purlFilePath = flag.String("f", "", "file containing list of package URLs")
	downloadDir  = flag.String("d", "", "directory to store downloaded tarballs")
)

// cmdError is a simple string error type, used when command usage
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

func downloadPackage(purl packageurl.PackageURL, dir string) error {
	pkg, err := worker.ResolvePurl(purl)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] %s@%s", pkg.EcosystemName(), pkg.Name(), pkg.Version())

	if downloadPath, err := pkg.Manager().DownloadArchive(pkg.Name(), pkg.Version(), dir); err != nil {
		fmt.Println()
		return err
	} else {
		fmt.Printf(" -> %s\n", downloadPath)
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

func processFileLine(text string) error {
	trimmed := strings.TrimSpace(text)
	if len(trimmed) == 0 || trimmed[0] == '#' {
		return nil
	}

	if purl, err := packageurl.FromString(trimmed); err != nil {
		return fmt.Errorf("invalid purl '%s': %w", text, err)
	} else if err := downloadPackage(purl, *downloadDir); err != nil {
		return fmt.Errorf("could not download %s: %w", text, err)
	}

	return nil
}

func run() error {
	flag.Parse()

	http.DefaultTransport = useragent.DefaultRoundTripper(http.DefaultTransport, "")

	if *purlFilePath == "" {
		return newCmdError("Please specify packages to download using -f <file>")
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

	defer purlFile.Close()

	scanner := bufio.NewScanner(purlFile)
	for line := 1; scanner.Scan(); line += 1 {
		if err := processFileLine(scanner.Text()); err != nil {
			fmt.Fprintf(os.Stderr, "line %d: %v\n", line, err)
		}
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
