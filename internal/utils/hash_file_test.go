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
		name        string
		contents    string
		prependType bool
		want        string
	}{
		{
			name:        "empty file",
			contents:    hashPairs[0][0],
			prependType: true,
			want:        "sha256:" + hashPairs[0][1],
		},
		{
			name:        "single line",
			contents:    hashPairs[1][0],
			prependType: true,
			want:        "sha256:" + hashPairs[1][1],
		},
		{
			name:        "single line hash only",
			contents:    hashPairs[1][0],
			prependType: false,
			want:        hashPairs[1][1],
		},
		{
			name:        "multi line",
			contents:    hashPairs[2][0],
			prependType: true,
			want:        "sha256:" + hashPairs[2][1],
		},
		{
			name:        "trailing new line",
			contents:    hashPairs[3][0],
			prependType: true,
			want:        "sha256:" + hashPairs[3][1],
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := filepath.Join(t.TempDir(), "file.txt")
			err := os.WriteFile(f, []byte(test.contents), 0o666)
			if err != nil {
				t.Fatalf("Failed to prepare hash file: %v", err)
			}
			got, err := utils.HashFile(f, test.prependType)
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
	got, err := utils.HashFile(f, true)
	if err == nil {
		t.Error("HashFile() returned no error; want an error")
	}
	if got != "" {
		t.Errorf("HashFile() = %v; want ''", got)
	}
}

func TestBasenameWithHash(t *testing.T) {
	tests := []struct {
		name             string
		filename         string
		fileContents     string
		prefix           string
		suffix           string
		expectedFilename string
	}{
		{
			name:             "empty file, no extension",
			filename:         "empty",
			fileContents:     hashPairs[0][0],
			expectedFilename: "empty" + hashPairs[0][1],
		},
		{
			name:             "empty file, with extension",
			filename:         "empty.txt",
			fileContents:     hashPairs[0][0],
			expectedFilename: "empty" + hashPairs[0][1] + ".txt",
		},
		{
			name:             "single line hash only, no extension",
			filename:         "single",
			fileContents:     hashPairs[1][0],
			expectedFilename: "single" + hashPairs[1][1],
		},
		{
			name:             "single line with prefix and suffix filename",
			filename:         "single.txt",
			fileContents:     hashPairs[1][0],
			prefix:           "-",
			suffix:           "xxxx",
			expectedFilename: "single-" + hashPairs[1][1] + "xxxx.txt",
		},
		{
			name:             "multi line, double file extension",
			filename:         "multi.txt.bak",
			fileContents:     hashPairs[1][0],
			expectedFilename: "multi" + hashPairs[1][1] + ".txt.bak",
		},
		{
			name:             "multi line, double file extension, no filename",
			filename:         ".txt.bak",
			fileContents:     hashPairs[1][0],
			prefix:           "test",
			suffix:           "xxxx",
			expectedFilename: "test" + hashPairs[1][1] + "xxxx.txt.bak",
		},
	}
	dir := t.TempDir()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, tt.filename)
			if err := os.WriteFile(path, []byte(tt.fileContents), 0o666); err != nil {
				t.Fatalf("Failed to create hash file: %v", err)
			}
			gotPath, err := utils.BasenameWithHash(path, tt.prefix, tt.suffix)
			if err != nil {
				t.Errorf("BasenameWithHash() error: %v", err)
				return
			}
			expectedPath := filepath.Join(dir, tt.expectedFilename)
			if gotPath != expectedPath {
				t.Errorf("BasenameWithHash() expected filename = %v but got %v", expectedPath, gotPath)
			}
		})
	}
}
