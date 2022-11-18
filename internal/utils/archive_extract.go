package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/ossf/package-analysis/internal/log"
)

// ExtractTarGzFile extracts a .tar.gz / .tgz file located at path, optionally
// accompanied by a temporary change to the given directory during extraction.
func ExtractTarGzFile(path string, chdir string) error {
	return processGzipFile(path, func(reader io.Reader) error {
		return extractTar(reader, chdir)
	})
}

func processGzipFile(path string, process func(io.Reader) error) error {
	gzFile, err := os.Open(path)
	if err != nil {
		return err
	}

	var unzippedBytes *gzip.Reader
	if unzippedBytes, err = gzip.NewReader(gzFile); err != nil {
		return err
	}

	defer func() {
		if err = unzippedBytes.Close(); err != nil {
			log.Error("failed to close gzip reader: %w", err)
		}
	}()

	if err = process(unzippedBytes); err != nil {
		return err
	}

	return nil
}

// extractTar extracts the contents of the given stream of bytes of a tar archive.
// If chdir is non-empty, the working directory will be temporarily changed to the
// specified one before the extraction takes place, so that relative paths are resolved
// relative to that directory. Otherwise, relative paths will be resolved relative to
// the current working directory.
func extractTar(tarStream io.Reader, chdir string) error {
	var header *tar.Header
	var err error

	if chdir != "" {
		var originalDir string
		if originalDir, err = os.Getwd(); err != nil {
			return fmt.Errorf("cannot get working directory: %w", err)
		}
		if err = os.Chdir(chdir); err != nil {
			return fmt.Errorf("cannot change working directory to %s: %w", chdir, err)
		}

		defer func() {
			if chDirErr := os.Chdir(originalDir); chDirErr != nil {
				log.Error("failed to restore working directory", "original dir", originalDir, "error", err)
			}
		}()
	}

	tarReader := tar.NewReader(tarStream)
	for header, err = tarReader.Next(); err == nil; header, err = tarReader.Next() {
		switch header.Typeflag {
		case tar.TypeDir:
			// make dir only readable by current user
			if err = os.Mkdir(header.Name, 0700); err != nil {
				return fmt.Errorf("mkdir failed: %w", err)
			}
		case tar.TypeReg:
			var extractedFile *os.File
			if extractedFile, err = os.Create(header.Name); err != nil {
				return fmt.Errorf("create failed: %w", err)
			}

			if _, err = io.Copy(extractedFile, tarReader); err != nil {
				if closeErr := extractedFile.Close(); closeErr != nil {
					return fmt.Errorf("copy failed: %w; close also failed: %v", err, closeErr)
				}
				return fmt.Errorf("copy failed: %w", err)
			}
			if err = extractedFile.Close(); err != nil {
				return fmt.Errorf("close failed: %w", err)
			}
		default:
			return fmt.Errorf("%s has unknown type %b", header.Name, header.Typeflag)
		}
	}

	if err != io.EOF {
		return fmt.Errorf("failed to read all archive entries: %w", err)
	}

	return nil
}
