package utils

import (
	"fmt"
	"os"
)

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
	defer f.Close()
	if err != nil {
		return "", err
	}
	_, err = f.Write(data)
	if err != nil {
		return "", err
	}
	f.Sync()
	return f.Name(), nil
}
