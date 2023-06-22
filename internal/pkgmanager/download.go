package pkgmanager

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

/*
downloadToPath creates (and/or truncates) a file at the given path, then writes
contents of whatever is at the given URL to that given file using downloadToFile,
and finally closes the file.

If any error occurs, the created file is removed.
*/
func downloadToPath(path string, url string) error {
	if path == "" {
		panic("path is empty")
	}
	if url == "" {
		panic("url is empty")
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}

	if downloadErr := downloadToFile(file, url); downloadErr != nil {
		// cleanup file
		removeErr := os.Remove(path)
		return errors.Join(downloadErr, removeErr)
	}

	if closeErr := file.Close(); closeErr != nil {
		// cleanup file
		removeErr := os.Remove(path)
		return errors.Join(closeErr, removeErr)
	}

	return nil
}

/*
downloadToFile writes the contents of whatever is at the given URL to the
given file, without opening or closing the file. If any errors occur while
making the network request, then no file operations will be performed.
*/
func downloadToFile(dest *os.File, url string) error {
	if url == "" {
		panic("url is empty")
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %s", resp.Status)
	}

	if _, err := io.Copy(dest, resp.Body); err != nil {
		return err
	}

	return nil
}
