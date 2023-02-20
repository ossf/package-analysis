package pkgecosystem

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

/*
downloadToDirectory downloads a file from the given URL to the given directory. The name of the
local file is obtained from the last element of the URL path when split on the '/' character.
On successful download, the full path to the downloaded file is returned.
*/
func downloadToDirectory(dir string, url string) (string, error) {
	fileName := path.Base(url)
	filePath := filepath.Join(dir, fileName)
	err := downloadToPath(filePath, url)
	if err != nil {
		return "", err
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
