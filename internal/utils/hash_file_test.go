package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ossf/package-analysis/internal/utils"
)

// pairs of strings and their SHA256 hash digests
var hashPairs = [][2]string{
	{"", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
	{"Hello, World!", "dffd6021bb2bd5b0af676290809ec3a53191dd81c7f70a4b28688a362182986f"},
	{"Hello,\nWorld!", "d62b51d504f02642dab5003959af0c1557094c7d49dcc544aba37a0a5d8d1d0d"},
	{"Hello,\nWorld!\n", "f5651768767f5e83d7001136251b6558a6d01550b04e12c1678ea3a0ca1e8a30"},
}

func TestHashFile(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{
			name:     "empty file",
			contents: hashPairs[0][0],
			want:     hashPairs[0][1],
		},
		{
			name:     "single line",
			contents: hashPairs[1][0],
			want:     hashPairs[1][1],
		},
		{
			name:     "multi line",
			contents: hashPairs[2][0],
			want:     hashPairs[2][1],
		},
		{
			name:     "trailing new line",
			contents: hashPairs[3][0],
			want:     hashPairs[3][1],
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := filepath.Join(t.TempDir(), "file.txt")
			err := os.WriteFile(f, []byte(test.contents), 0o666)
			if err != nil {
				t.Fatalf("Failed to prepare hash file: %v", err)
			}
			got, err := utils.SHA256Hash(f)
			if err != nil {
				t.Fatalf("Failed to generate hash: %v", err)
			}
			if got != test.want {
				t.Errorf("SHA256Hash() = %v; want %v", got, test.want)
			}
		})
	}
}

func TestHashFile_MissingFile(t *testing.T) {
	f := filepath.Join(t.TempDir(), "missing.txt")
	got, err := utils.SHA256Hash(f)
	if err == nil {
		t.Error("SHA256Hash() returned no error; want an error")
	}
	if got != "" {
		t.Errorf("SHA256Hash() = %v; want ''", got)
	}
}
