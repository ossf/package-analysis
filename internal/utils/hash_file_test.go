package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ossf/package-analysis/internal/utils"
)

func TestHashFile(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{
			name:     "empty file",
			contents: "",
			want:     "sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:     "single line",
			contents: "Hello, World!",
			want:     "sha256:dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f",
		},
		{
			name:     "mutli line",
			contents: "Hello,\nWorld!",
			want:     "sha256:d62b51d504f02642dab5003959af0c1557094c7d49dcc544aba37a0a5d8d1d0d",
		},
		{
			name:     "trailing new line",
			contents: "Hello,\nWorld!\n",
			want:     "sha256:f5651768767f5e83d7001136251b6558a6d01550b04e12c1678ea3a0ca1e8a30",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := filepath.Join(t.TempDir(), "file.txt")
			err := os.WriteFile(f, []byte(test.contents), 0o666)
			if err != nil {
				t.Fatalf("Failed to prepare hash file: %v", err)
			}
			got, err := utils.HashFile(f)
			if err != nil {
				t.Fatalf("Failed to generate hash: %v", err)
			}
			if got != test.want {
				t.Errorf("HashFile() = %v; want %v", got, test.want)
			}
		})
	}
}

func TestHashFile_MissingFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "missing.txt")
	got, err := utils.HashFile(f)
	if err == nil {
		t.Error("HashFile() returned no error; want an error")
	}
	if got != "" {
		t.Errorf("HashFile() = %v; want ''", got)
	}
}
