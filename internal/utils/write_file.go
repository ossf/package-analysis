package utils

import (
	"fmt"
	"os"
	"path/filepath"
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
			return fmt.Errorf("could not set exec permissions on %s: %v", path, err)
		}
	}

	return nil
}

/* Writes a a file and flushes the buffer */
func CreateAndWriteTempFile(fileName string, data []byte) (string, error) {
	dirPath := filepath.Join(os.TempDir(), tempFolder)
	os.Mkdir(dirPath, 0666)
	f, err := os.CreateTemp(dirPath, fileName)
	if err != nil {
		return "", err
	}
	f.Write(data)
	f.Close()
	f.Sync()
	return f.Name(), nil

}

// change return type to rune?
func ReadTempFile(fileName string) ([]byte, error) {
	// permissions?
	f, err := os.OpenFile(fileName, os.O_RDWR, 0666)
	if err != nil {
		return []byte{}, err
	}
	// Seek to the beginning of the file
	f.Seek(0, 0)
	fileContents, err := os.ReadFile(fileName)
	f.Close()
	return fileContents, nil
}

func CleanUpTempFiles() {
	dirPath := filepath.Join(os.TempDir(), tempFolder)
	os.RemoveAll(dirPath)
}
