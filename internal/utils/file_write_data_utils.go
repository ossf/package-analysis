package utils

import (
	"os"
	"path/filepath"
)

/*
Subfolder where write buffer data will be saved to disk before uploaded to a cloud bucket.
This subfolder needs to be shared across files so all functions that access it will be defined here.
*/

const writeBufferFolder = "worker_tmp/write_buffers"

// CreateAndWriteTempFile writes a file in the directory specified by
// writeBufferFolder.
//
// This directory must be cleaned up with a call to RemoveTempFilesDirectory().
func CreateAndWriteTempFile(fileName string, data []byte) error {
	err := os.MkdirAll(writeBufferFolder, 0777)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(writeBufferFolder, fileName))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func OpenTempFile(fileName string) (*os.File, error) {
	return os.Open(filepath.Join(writeBufferFolder, fileName))
}

func RemoveTempFilesDirectory() error {
	return os.RemoveAll(writeBufferFolder)
}
