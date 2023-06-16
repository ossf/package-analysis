package pkgmanager

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ossf/package-analysis/internal/utils"
)

/*
downloadToDirectory downloads a file from the given URL to the given directory.
On successful download, the full path to the downloaded file is returned.

fileName argument is required with no default value, and the hash of the file
is appended to the given filename. If an error occurs during hashing, then the
original filename is used.
*/
func downloadToDirectory(dir string, url string, fileName string) (string, error) {
	filePath := filepath.Join(dir, fileName)

	if err := downloadToPath(filePath, url); err != nil {
		return "", err
	}

	hash, err := utils.HashFile(filePath)
	if err != nil {
		// Error is ignored and the path with the original filename is returned
		// We treat hashing as 'best-effort' rather than a strictly necessary part of the function
		return filePath, nil
	}

	hashedFilePath := strings.Join([]string{filePath, hash[7:]}, "-")

	err = os.Rename(filePath, hashedFilePath)
	if err != nil {
		// See above comment
		return filePath, nil
	}

	return hashedFilePath, nil
}

/*
downloadToPath creates (and/or truncates) a file at the given path, then writes
contents of whatever is at the given URL to that given file using downloadToFile,
and finally closes the file.
*/
func downloadToPath(path string, url string) (err error) {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer func() {
		closeErr := file.Close()
		if err == nil {
			err = closeErr
		} else if closeErr != nil {
			err = fmt.Errorf("%w; error closing file %s: %v", err, path, closeErr)
		}
	}()

	err = downloadToFile(file, url)

	if err != nil {
		return err
	}

	return nil
}

/*
downloadToFile writes the contents of whatever is at the given URL to the
given file, without opening or closing the file. If any errors occur while
making the network request, then no file operations will be performed.
*/
func downloadToFile(dest *os.File, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %s", resp.Status)
	}

	_, err = io.Copy(dest, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
