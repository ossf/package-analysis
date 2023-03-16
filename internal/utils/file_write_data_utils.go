package utils

import (
	"os"
	"path/filepath"
)

/*
Directory where write buffer data will be saved to disk before uploaded to a cloud bucket.
This directory needs to be shared across files so all functions that access it will be defined here.
*/
const temp_write_buf_dir = "tmp/temp_write_buffers"

/* Writes a file in the directory specified by temp_write_buf_dir and flushes the buffer */
func CreateAndWriteTempFile(fileName string, data []byte) error {
	err := os.Mkdir(temp_write_buf_dir, 0666)
	f, err := os.Create(filepath.Join(temp_write_buf_dir, fileName))
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
	return os.Open(filepath.Join(temp_write_buf_dir, fileName))
}

func RemoveTempFilesDirectory() error {
	return os.RemoveAll(temp_write_buf_dir)
}
