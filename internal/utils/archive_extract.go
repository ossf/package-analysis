package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ossf/package-analysis/internal/log"
)

// ExtractTarGzFile extracts a .tar.gz / .tgz file located at filePath,
// using outputDir as the root of the extracted files.
func ExtractTarGzFile(filePath string, outputDir string) error {
	return processGzipFile(filePath, func(reader io.Reader) error {
		return extractTar(reader, outputDir)
	})
}

func processGzipFile(filePath string, process func(io.Reader) error) error {
	gzFile, err := os.Open(filePath)
	if err != nil {
		return err
	}

	var unzippedBytes *gzip.Reader
	if unzippedBytes, err = gzip.NewReader(gzFile); err != nil {
		return err
	}

	defer func() {
		if closeErr := unzippedBytes.Close(); closeErr != nil {
			log.Error("failed to close gzip reader", "error", closeErr)
		}
	}()

	if err := process(unzippedBytes); err != nil {
		return err
	}
	return nil
}

/*
extractTar extracts the contents of the given stream of bytes of a tar archive, using
outputDir as the root of the extracted files.
*/
func extractTar(tarStream io.Reader, outputDir string) error {
	if outputDir == "" {
		return fmt.Errorf("outputDir is empty")
	}

	tarReader := tar.NewReader(tarStream)

	var header *tar.Header
	var err error
	for header, err = tarReader.Next(); err == nil; header, err = tarReader.Next() {
		outputPath := filepath.Join(outputDir, header.Name)
		// check for ZipSlip (https://snyk.io/research/zip-slip-vulnerability) by ensuring
		// outputPath (cleaned) actually is inside output directory that was specified
		if !strings.HasPrefix(outputPath, filepath.Join(outputDir)+string(os.PathSeparator)) {
			// Note: this error string is used in a test
			return fmt.Errorf("archive path escapes output dir: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// make dir only readable by current user
			if err = os.Mkdir(outputPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("mkdir failed: %w", err)
			}
		case tar.TypeReg:
			fileInfo := header.FileInfo()
			openFlags := os.O_RDWR | os.O_CREATE | os.O_TRUNC // copied from os.Create()

			// ensure containing directories exist; some archives don't include an explicit entry
			// for parent directories
			parentDir := filepath.Dir(outputPath)
			if err = os.MkdirAll(parentDir, 0o755); err != nil {
				return fmt.Errorf("create parent dirs for %s failed: %w", header.Name, err)
			}

			var extractedFile *os.File
			if extractedFile, err = os.OpenFile(outputPath, openFlags, fileInfo.Mode()); err != nil {
				return fmt.Errorf("create file failed: %w", err)
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
