package utils

import (
	"os"
	"path/filepath"
)

/*
Subfolder where write buffer data will be saved to disk before uploaded to a cloud bucket.
This subfolder needs to be shared across files so all functions that access it will be defined here.
*/

const write_buffer_folder = "temp_write_buffers"

/*
Writes a file in the directory specified by write_buffer_folder and flushes the buffer.
This directory is meant to be cleaned up through the RemoveTempFilesDirectory() method.
*/
func CreateAndWriteTempFile(fileName string, data []byte) error {
	err := os.MkdirAll(write_buffer_folder, 0777)
	if err != nil {
		return err
	}

	f, err := os.Create(filepath.Join(write_buffer_folder, fileName))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	f.Sync()
	return nil
}

func OpenTempFile(fileName string) (*os.File, error) {
	return os.Open(filepath.Join(write_buffer_folder, fileName))
}

func RemoveTempFilesDirectory() error {
	return os.RemoveAll(write_buffer_folder)
}
