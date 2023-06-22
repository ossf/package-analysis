package pkgmanager

import (
	"os"
	"testing"

	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

type downloadTestSpec struct {
	name       string
	pkgName    string
	pkgVersion string
	wantErr    bool
}

func testDownload(t *testing.T, tests []downloadTestSpec, manager *PkgManager) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downloadDir := t.TempDir()
			downloadPath, err := manager.DownloadArchive(tt.pkgName, tt.pkgVersion, downloadDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Want error: %v; got error: %v", tt.wantErr, err)
				return
			}

			if err != nil {
				// File wasn't meant to download properly
				return
			}

			if err := os.Remove(downloadPath); err != nil {
				t.Errorf("Error removing file: %v", err)
			}
		})
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

func TestDownloadAndHashCheck(t *testing.T) {
	tests := []struct {
		name         string
		pkgEcosystem pkgecosystem.Ecosystem
		pkgName      string
		pkgVersion   string
		archiveHash  string
		wantErr      bool
	}{
		{
			name:         "pypi black 23.3.0",
			pkgEcosystem: pkgecosystem.PyPI,
			pkgName:      "black",
			pkgVersion:   "23.3.0",
			archiveHash:  "1c7b8d606e728a41ea1ccbd7264677e494e87cf630e399262ced92d4a8dac940",
			wantErr:      false,
		},
		{
			name:         "npm invalid package name",
			pkgEcosystem: pkgecosystem.PyPI,
			pkgName:      "3i3ii3ii3i3i",
			pkgVersion:   "",
			archiveHash:  "",
			wantErr:      true,
		},
		{
			name:         "pypi black invalid version",
			pkgEcosystem: pkgecosystem.PyPI,
			pkgName:      "black",
			pkgVersion:   "23333.3.0",
			archiveHash:  "",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			downloadDir := t.TempDir()
			downloadPath, err := Manager(tt.pkgEcosystem).DownloadArchive(tt.pkgName, tt.pkgVersion, downloadDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("Want error: %v; got error: %v", tt.wantErr, err)
				return
			}
			if err != nil {
				// File wasn't meant to download properly
				return
			}

			gotHash, err := utils.HashFile(downloadPath, false)
			if err != nil {
				// hashing isn't meant to throw an error
				t.Errorf("hashing failed: %v", err)
				return
			}

			if tt.archiveHash != gotHash {
				t.Errorf("Expected hash %s, got %s", tt.archiveHash, gotHash)
			}

			if err := os.Remove(downloadPath); err != nil {
				t.Errorf("Error removing downloaded archive: %v", err)
			}
		})
	}
}
