package basicdata

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/pkg/valuecounts"
)

type testFile struct {
	filename     string
	contents     []byte
	contentsHash string
	fileType     string
	lineLengths  valuecounts.ValueCounts
}

var testFiles = []testFile{
	{
		filename:     "test1.txt",
		contents:     []byte("hello test 1!\n"),
		contentsHash: "bd96959573979235b87180b0b7513c7f1d5cbf046b263f366f2f10fe1b966494",
		fileType:     "ASCII text",
		lineLengths:  valuecounts.Count([]int{13}),
	},
	{
		filename:     "test2.txt",
		contents:     []byte("#! /bin/bash\necho 'Hello test 2'\n"),
		contentsHash: "6179db3c673ceddcdbd384116ae4d301d64e65fc2686db9ba64945677a5a893c",
		fileType:     "Bourne-Again shell script, ASCII text executable",
		lineLengths:  valuecounts.Count([]int{12, 19}),
	},
}

func TestGetBasicData(t *testing.T) {
	tests := []struct {
		name    string
		files   []testFile
		wantErr bool
	}{
		{
			name:    "test no files",
			files:   nil,
			wantErr: false,
		},
		{
			name:    "test one file",
			files:   []testFile{testFiles[0]},
			wantErr: false,
		},
		{
			name:    "test two files",
			files:   []testFile{testFiles[0], testFiles[1]},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := t.TempDir()
			paths := utils.Transform(tt.files, func(f testFile) string {
				return filepath.Join(testDir, f.filename)
			})

			for i := range tt.files {
				if err := os.WriteFile(paths[i], tt.files[i].contents, 0o666); err != nil {
					t.Fatalf("failed to write test file %d: %v", i, err)
				}
			}

			got, err := Analyze(context.Background(), paths)
			if (err != nil) != tt.wantErr {
				t.Errorf("detectFileTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			wantData := utils.Transform(tt.files, func(f testFile) FileData {
				return FileData{
					DetectedType: f.fileType,
					Size:         int64(len(f.contents)),
					SHA256:       f.contentsHash,
					LineLengths:  f.lineLengths,
				}
			})

			if !reflect.DeepEqual(got, wantData) {
				t.Errorf("TestGetBasicData() data mismatch:\n"+
					"== got == \n%v\n== want ==\n%v", got, wantData)
			}
		})
	}
}
