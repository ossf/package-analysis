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

/* Writes a file in the directory specified by temp_write_buf_dir and flushes the buffer */
func CreateAndWriteTempFile(fileName string, data []byte) error {
	tempWriteBufDir := filepath.Join(os.TempDir(), write_buffer_folder)
	err := os.Mkdir(tempWriteBufDir, 0777)
	f, err := os.Create(filepath.Join(tempWriteBufDir, fileName))
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
	return os.Open(filepath.Join(os.TempDir(), write_buffer_folder, fileName))
}

func RemoveTempFilesDirectory() error {
	return os.RemoveAll(filepath.Join(os.TempDir(), write_buffer_folder))
}
