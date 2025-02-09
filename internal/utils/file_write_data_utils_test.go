package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true
	}
	return false
}

func TestCreateAndWriteTempFile(t *testing.T) {
	// Cleanup any old test runs
	RemoveTempFilesDirectory()

	fileName := "CreateAndWriteTempFile_testfile.txt"
	filePath := filepath.Join(writeBufferFolder, fileName)

	defer func() {
		err := RemoveTempFilesDirectory()
		if err != nil {
			t.Logf("%s could not be cleaned up: %s.", fileName, err)
		}
	}()

	CreateAndWriteTempFile(fileName, []byte("This test file is safe to remove."))

	if !fileExists(filePath) {
		t.Errorf("CreateAndWriteTempFile(): did not create file, want %v", fileName)
	}
}
