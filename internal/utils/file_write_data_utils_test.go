package utils

import (
	"os"
	"path/filepath"
	"reflect"
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

	defer RemoveTempFilesDirectory()

	CreateAndWriteTempFile(fileName, []byte("This test file is safe to remove."))

	if !fileExists(filePath) {
		t.Errorf("CreateAndWriteTempFile(): did not create file, want %v", fileName)
	}
}

func TestOpenTempFile(t *testing.T) {
	fileName := "CreateAndWriteTempFile_testfile.txt"
	fileBody := []byte("This test file is safe to remove.")
	CreateAndWriteTempFile(fileName, []byte(fileBody))

	defer RemoveTempFilesDirectory()

	file, err := OpenTempFile(fileName)
	if err != nil {
		t.Errorf("%s could not be opened for test", fileName)
	}
	defer file.Close()

	actualName := file.Name()
	if actualName != filepath.Join(writeBufferFolder, fileName) {
		t.Errorf("Name() = %s; want %s", actualName, fileName)
	}

	actualBody := []byte{}
	_, err = file.Read(actualBody)
	if err != nil {
		t.Errorf("%s could not be read for test", fileName)
	}
	if !reflect.DeepEqual(actualBody, fileBody) {
		t.Errorf("Read() = %s; want %s", actualBody, fileBody)
	}
}

func TestRemoveTempFilesDirectory(t *testing.T) {
	if err := os.MkdirAll(writeBufferFolder, 0777); err != nil {
		t.Errorf("%s could not be created for test", writeBufferFolder)
	}

	err := RemoveTempFilesDirectory()
	if err != nil {
		t.Errorf("Error removing temp folder: %s", err)
	}

	if fileExists((writeBufferFolder)) {
		t.Errorf("RemoveTempFilesDirectory(): folder exists, want no folder.")
	}
}
