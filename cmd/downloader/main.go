package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/package-url/packageurl-go"

	"github.com/ossf/package-analysis/internal/pkgmanager"
	"github.com/ossf/package-analysis/internal/worker"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

// Command-line tool to download a list of package archives, specified by purl
// See https://github.com/package-url/purl-spec
var (
	packagesFile = flag.String("f", "", "file containing packages to download")
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
	var ecosystem pkgecosystem.Ecosystem
	if err := ecosystem.UnmarshalText([]byte(purl.Type)); err != nil {
		return fmt.Errorf("unsupported package ecosystem for purl %s: %s", purl.String(), purl.Type)
	}

	manager := pkgmanager.Manager(ecosystem)
	if manager == nil {
		return fmt.Errorf("unsupported package ecosystem for purl %s: %s", purl.String(), purl.Type)
	}

	// Prepend package namespace to package name, if present
	var pkgName string
	if purl.Namespace != "" {
		pkgName = purl.Namespace + "/" + purl.Name
	} else {
		pkgName = purl.Name
	}

	// Get the latest package version if not specified in the purl
	pkg, err := worker.ResolvePkg(manager, pkgName, purl.Version, "")
	if err != nil {
		return err
	}

	fmt.Printf("[%s] %s@%s\n", ecosystem, pkg.Name(), pkg.Version())

	if _, err := manager.DownloadArchive(pkg.Name(), pkg.Version(), downloadDir); err != nil {
		return err
	}

	return nil
}

func run() error {
	flag.Parse()

	if *packagesFile == "" {
		return newCmdError("package list file (-f) is a required argument")
	}
	if *downloadDir == "" {
		*downloadDir = "."
	}

	if stat, err := os.Stat(*downloadDir); err != nil {
		return fmt.Errorf("could not stat download directory %s: %w", downloadDir, err)
	} else if !stat.IsDir() {
		return fmt.Errorf("%s is not a directory", *downloadDir)
	}

	var fileLines []string
	if purlBytes, err := os.ReadFile(*packagesFile); err != nil {
		return err
	} else {
		fileLines = strings.Split(string(purlBytes), "\n")
	}

	var purls []packageurl.PackageURL
	for i, s := range fileLines {
		trimmed := strings.TrimSpace(s)
		if len(trimmed) == 0 || trimmed[0] == '#' {
			continue
		}
		if purl, err := packageurl.FromString(trimmed); err != nil {
			return fmt.Errorf("could not parse purl '%s' on line %d: %w", s, i+1, err)
		} else {
			purls = append(purls, purl)
		}
	}

	for _, purl := range purls {
		if err := downloadPackage(purl, *downloadDir); err != nil {
			return err
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
