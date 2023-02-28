package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"

	"github.com/ossf/package-analysis/internal/log"
)

const tempFolder = "temp_writes_folder"

/*
WriteFile writes the given file contents to the given path.
The file may optionally be marked as executable.
*/
func WriteFile(path string, contents []byte, executable bool) error {
	if err := os.WriteFile(path, contents, 0o666); err != nil {
		return err
	}

	if executable {
		if err := os.Chmod(path, 0o777); err != nil {
			return fmt.Errorf("could not set exec permissions on %s: %w", path, err)
		}
	}

	return nil
}

/* Writes a temp file and flushes the buffer */
func CreateAndWriteTempFile(fileName string, data []byte) (string, error) {
	f, err := os.CreateTemp("", fileName)
	if err != nil {
		return "", err
	}
	_, writeErr := f.Write(data)
	if writeErr != nil {
		return "", writeErr
	}
	f.Close()
	f.Sync()
	return f.Name(), nil

}

/* Reads the temp file at fileName and removes the file afterwards */
func ReadAndRemoveTempFile(fileName string) ([]byte, error) {
	f, err := os.OpenFile(fileName, os.O_RDWR, 0666)
	if err != nil {
		return []byte{}, err
	}
	// Seek to the beginning of the file
	//f.Seek(0, 0)
	fileContents, err := os.ReadFile(fileName)
	f.Close()
	os.Remove(fileName)
	return fileContents, err
}

func WriteFilesToZip(writeBufferPaths []string, zipFile *os.File) {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	log.Debug("mem stats before write files to zip")
	log.Debug(strconv.FormatUint(rtm.Alloc, 10))
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	for _, path := range writeBufferPaths {
		file, err := os.Open(path)
		if err != nil {
			return
		}
		fileContents, err := os.ReadFile(path)
		writeBufferId := GetSHA256Hash(fileContents)
		w, err := zipWriter.Create(writeBufferId)
		if err != nil {
			return
		}
		if _, err := io.Copy(w, file); err != nil {
			return
		}
		// do we need to flush the writer?
		// figure out how to defer this close in the for loop
		os.Remove(path)
		file.Close()
	}
	var rtm2 runtime.MemStats
	runtime.ReadMemStats(&rtm2)
	log.Debug("mem stats after  write files to zip")
	log.Debug(strconv.FormatUint(rtm2.Alloc, 10))

}
