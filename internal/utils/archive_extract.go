package utils

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ExtractArchiveFile extracts a .tar.gz / .tgz file or .zip file located
// at archivePath, using outputDir as the root of the extracted files.
func ExtractArchiveFile(archivePath string, outputDir string) error {
	if outputDir == "" {
		return fmt.Errorf("outputDir is empty")
	}

	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	fileType, err := detectFileType(f)
	if err != nil {
		return err
	}

	if fileType == "application/zip" {
		return processZipFile(archivePath, outputDir)
	} else if fileType == "application/x-gzip" {
		return processGzipFile(f, func(reader io.Reader) error {
			return extractTar(reader, outputDir)
		})
	} else {
		return fmt.Errorf("%s is not a supported archive file type: %s", archivePath, fileType)
	}
}

func processZipFile(filePath, outputDir string) error {
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		return fmt.Errorf("create dir for %s failed: %w", outputDir, err)
	}

	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer zipReader.Close()

	for _, file := range zipReader.File {
		outputPath := filepath.Join(outputDir, file.Name)

		// check for ZipSlip (https://snyk.io/research/zip-slip-vulnerability) by ensuring
		// outputPath (cleaned) actually is inside output directory that was specified
		if !strings.HasPrefix(outputPath, filepath.Join(outputDir)+string(os.PathSeparator)) {
			// Note: this error string is used in a test
			return fmt.Errorf("archive path escapes output dir: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(outputPath, 0o755); err != nil {
				return err
			}
			continue
		}

		// Ensure parent directories exist before creating the file
		if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
			return err
		}

		srcFile, err := file.Open()
		if err != nil {
			return err
		}
		defer srcFile.Close()

		destFile, err := os.Create(outputPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func processGzipFile(gzFile *os.File, process func(io.Reader) error) error {
	unzippedBytes, err := gzip.NewReader(gzFile)
	if err != nil {
		return err
	}

	defer func() {
		if closeErr := unzippedBytes.Close(); closeErr != nil {
			slog.Error("failed to close gzip reader", "error", closeErr)
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

func detectFileType(archiveFile *os.File) (string, error) {
	// DetectContentType never uses more than the first 512 bytes.
	buffer := make([]byte, 512)
	_, err := archiveFile.Read(buffer)
	if err != nil {
		return "", err
	}

	mimeType := http.DetectContentType(buffer)

	return mimeType, nil
}
