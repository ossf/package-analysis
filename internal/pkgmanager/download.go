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

fileName argument allows passing a custom filename for local file.
The default filename will be obtained from the last element
of the URL path when split on the '/' character.
*/
func downloadToDirectory(dir string, url string, fileName string) (string, error) {
	filePath := filepath.Join(dir, fileName)
	err := downloadToPath(filePath, url)
	if err != nil {
		return "", err
	}

	hash, err := utils.HashFile(filePath)
	if err != nil {
		return filePath, nil
	}

	err = os.Rename(filePath, strings.Join([]string{filePath, hash[7:]}, "-"))
	if err != nil {
		return filePath, nil
	}

	return filePath, nil
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
