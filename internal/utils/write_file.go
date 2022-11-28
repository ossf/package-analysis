package utils

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
)

const defaultFileMode os.FileMode = 0666
const executableFlag os.FileMode = 0111

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

	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	if err == nil && executable {
		// file exists - make sure it's executable
		if err := os.Chmod(path, fileInfo.Mode().Perm()|executableFlag); err != nil {
			return fmt.Errorf("could not set executable bit on %s: %v", path, err)
		}
	}

	fileMode := defaultFileMode
	if executable {
		fileMode |= executableFlag
	}

	return os.WriteFile(path, contents, fileMode)
}
