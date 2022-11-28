package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

/*
WriteFile writes the given file contents to the given path.
The file may optionally be marked as executable.

In more detail:

 1. If a non-regular file exists at path, an error is returned.
 2. If a regular file exists at path, it is truncated and marked as executable if applicable.
 3. If no file exists at path, a new one is created with executable permissions if applicable

In cases 2 and 3 above, the contents of scriptSource are then written to the file.
*/
func WriteFile(path string, contents []byte, executable bool) (err error) {
	fileInfo, err := os.Stat(path)

	if err == nil {
		if !fileInfo.Mode().IsRegular() {
			return fmt.Errorf("not a regular file: %s", path)
		}
		if executable {
			// else it's a regular, empty file that we'll chmod
			if err = os.Chmod(path, fileInfo.Mode().Perm()|0100); err != nil {
				return fmt.Errorf("could not set executable bit on %s: %v", path, err)
			}
		}
	}

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("could not stat %s: %v", path, err)
	}

	var fileMode os.FileMode
	if executable {
		fileMode = 0700
	} else {
		fileMode = 0600
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return fmt.Errorf("could not open or create file %s: %v", path, err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil && err != nil {
			err = closeErr
		}
	}()

	if _, err = file.Write(contents); err != nil {
		return fmt.Errorf("could not write to %s: %v", path, err)
	}

	return
}
