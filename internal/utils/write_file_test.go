package utils

import (
	"bytes"
	"os"
	"testing"
)

// Test creating, reading, and deleting temp files.
func TestCreateReadAndDeleteTempFile(t *testing.T) {
	fileName := "test"
	data := []byte("Test data to be written.")
	tempFilePath, err := CreateAndWriteTempFile(fileName, data)
	if err != nil {
		t.Errorf("Could not create and write temp file %v", err)
	}
	bytesRead, err := ReadAndRemoveTempFile(tempFilePath)
	if err != nil {
		t.Errorf("Could not read and remove temp file %v", err)
	}
	if bytes.Compare(data, bytesRead) != 0 {
		t.Errorf("Bytes written and bytes read are not equal")
	}
	if _, err := os.Stat(tempFilePath); err == nil || !os.IsNotExist(err) {
		t.Errorf("File was not removed correctly.")
	}
}
