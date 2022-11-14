package pkgecosystem

import (
	"fmt"
	"os"
	"testing"
)

func TestDownloadNpmArchive(t *testing.T) {
	tests := []struct {
		name       string
		pkgName    string
		pkgVersion string
		wantErr    bool
	}{
		{"download 'abc' with latest version", "abc", "", false},
		{"download 'abc' with valid version", "abc", "0.4", false},
		{"download 'abc' with invalid version", "abc", "0.4333", true},
		{"download invalid package", "abcdddddddd", "", true},
	}
	tmpDir, err := os.MkdirTemp("", "package-analysis-test-npm-dl")
	fmt.Printf("Using temp dir %s\n", tmpDir)
	if err != nil {
		t.Fatalf("Could not create temp dir for downloads: %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, err := downloadNpmArchive(tt.pkgName, tt.pkgVersion, tmpDir)
			if err != nil {
				if !tt.wantErr {
					t.Error("Unexpected error")
				}
				fmt.Printf("%v\n", err)
			} else if gotPath == "" {
				t.Errorf("downloadNpmArchive() got empty path")
			}
		})
	}
	err = os.RemoveAll(tmpDir)
	if err != nil {
		t.Errorf("error removing temp dir: %v", err)
	}
}
