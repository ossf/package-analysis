package pkgmanager

import (
	"fmt"
	"os"
	"testing"

	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

type downloadTestSpec struct {
	name       string
	pkgName    string
	pkgVersion string
	wantErr    bool
}

func testDownload(t *testing.T, tests []downloadTestSpec, manager *PkgManager) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "package-analysis-test-npm-dl")
	if err != nil {
		t.Fatalf("Could not create temp dir for downloads: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downloadPath, err := manager.DownloadArchive(tt.pkgName, tt.pkgVersion, tmpDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Want error: %v; got error: %v", tt.wantErr, err)
				return
			}

			if err != nil {
				// File wasn't meant to download properly
				return
			}

			if downloadPath == "" {
				t.Errorf("downloadNPMArchive() returned no error but empty path")
				return
			}

			err = os.Remove(downloadPath)
			if err != nil {
				t.Errorf("Error removing file: %v", err)
			}
		})
	}

	err = os.RemoveAll(tmpDir)
	if err != nil {
		t.Errorf("error removing temp dir (%s): %v", tmpDir, err)
	}
}

func TestDownloadNpmArchive(t *testing.T) {
	tests := []downloadTestSpec{
		{"pkgname=async version='latest'", "async", "latest", false},
		{"pkgname=async version=(valid)", "async", "3.2.4", false},
		{"pkgname=async version=(invalid)", "async", "3.2.4444444", true},
		{"pkgname=(invalid)", "fr(2t5j923)", "latest", true},
	}

	testDownload(t, tests, Manager(pkgecosystem.NPM))
}

func TestDownloadPyPIArchive(t *testing.T) {
	tests := []downloadTestSpec{
		{"pkgname=urllib3 version=(valid)", "urllib3", "1.26.11", false},
		{"pkgname=urllib3 version=(invalid)", "urllib3", "1.26.111", true},
		{"pkgname=(invalid)", "fr(2t5j923)", "123", true},
	}

	testDownload(t, tests, Manager(pkgecosystem.PyPI))
}

func TestDownloadCratesArchive(t *testing.T) {
	tests := []downloadTestSpec{
		{"pkgname=rand version=(valid)", "rand", "0.8.5", false},
		{"pkgname=rand version=(invalid)", "rand", "0.8.55", true},
		{"pkgname=(invalid)", "fr(2t5j923)", "123", true},
	}

	testDownload(t, tests, Manager(pkgecosystem.CratesIO))
}

func TestDownloadToDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "package-analysis-test-dl")
	if err != nil {
		t.Fatalf("Could not create temp dir for downloads: %v", err)
	}

	testPkg := "black"
	version := "23.3.0"
	checksum := "1c7b8d606e728a41ea1ccbd7264677e494e87cf630e399262ced92d4a8dac940"
	fileName := fmt.Sprintf("%s-%s.tar.gz-%s", testPkg, version, checksum)

	testURL, err := getPyPIArchiveURL(testPkg, version)
	if err != nil {
		t.Fatalf("Could not get PyPI archive url")
	}

	tests := []struct {
		name     string
		url      string
		dir      string
		filename string
		wantErr  bool
	}{
		{testURL + " (plain)", testURL, tmpDir, fileName, false},
		{testURL + "123 (invalid URL)", testURL + "123", tmpDir, testURL + "123", true},
		{testURL + " (invalid dir)", testURL, "/tmp/does/not/exist", fileName, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downloadedFile, err := downloadToDirectory(tt.dir, tt.url, DefaultArchiveFilename(testPkg, version, testURL))
			if (err != nil) != tt.wantErr {
				t.Errorf("Want error: %v; got error: %v", tt.wantErr, err)
				return
			}

			if err != nil {
				// File wasn't meant to download properly
				return
			}

			stat, err := os.Stat(downloadedFile)
			if err != nil {
				t.Errorf("stat() returned error: %v", err)
				return
			}
			if stat.Name() != tt.filename {
				t.Errorf("Expected filename %s, got filename %s", tt.filename, stat.Name())
			}
			err = os.Remove(downloadedFile)
			if err != nil {
				t.Errorf("Error removing file: %v", err)
			}
		})
	}
	err = os.RemoveAll(tmpDir)
	if err != nil {
		t.Errorf("error removing temp dir (%s): %v", tmpDir, err)
	}
}
