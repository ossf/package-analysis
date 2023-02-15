package staticanalysis

import (
	"os"
	"reflect"
	"testing"
)

func TestGetFileTypes(t *testing.T) {
	testDir := t.TempDir()
	fileName1 := testDir + string(os.PathSeparator) + "test1.txt"
	fileName2 := testDir + string(os.PathSeparator) + "test2.txt"

	if err := os.WriteFile(fileName1, []byte("hello test 1!\n"), 0o666); err != nil {
		t.Fatalf("failed to write test file 1: %v", err)
	}
	if err := os.WriteFile(fileName2, []byte("#! /bin/bash\necho 'Hello test 2'\n"), 0o666); err != nil {
		t.Fatalf("failed to write test file 2: %v", err)
	}

	tests := []struct {
		name     string
		fileList []string
		want     []string
		wantErr  bool
	}{
		{
			name:     "test no files",
			fileList: []string{},
			want:     []string{},
			wantErr:  false,
		},
		{
			name:     "test one file",
			fileList: []string{fileName1},
			want:     []string{"ASCII text"},
			wantErr:  false,
		},
		{
			name:     "test two files",
			fileList: []string{fileName1, fileName2},
			want:     []string{"ASCII text", "Bourne-Again shell script, ASCII text executable"},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getFileTypes(tt.fileList)
			if (err != nil) != tt.wantErr {
				t.Errorf("getFileTypes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFileTypes() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}
