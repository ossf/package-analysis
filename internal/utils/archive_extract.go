package utils

import (
	"github.com/mholt/archiver/v4"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractArchiveFile extracts a .tar.gz / .tgz file located at archivePath,
// using outputDir as the root of the extracted files.
func ExtractArchiveFile(archivePath string, outputDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	format, input, err := archiver.Identify(archivePath, f)
	if err != nil {
		return err
	}

	return format.(archiver.Extractor).Extract(context.Background(), input, nil, func(ctx context.Context, f archiver.File) error {
		outputPath := filepath.Join(outputDir, f.NameInArchive)
		// check for ZipSlip (https://snyk.io/research/zip-slip-vulnerability) by ensuring
		// outputPath (cleaned) actually is inside output directory that was specified
		if !strings.HasPrefix(outputPath, filepath.Join(outputDir)+string(os.PathSeparator)) {
			// Note: this error string is used in a test
			return fmt.Errorf("archive path escapes output dir: %s", f.NameInArchive)
		}

		if f.IsDir() {
			// make dir only readable by current user
			if err = os.MkdirAll(outputPath, f.Mode()); err != nil {
				return fmt.Errorf("mkdir failed: %w", err)
			}
		} else {
			openFlags := os.O_RDWR | os.O_CREATE | os.O_TRUNC // copied from os.Create()

			// ensure containing directories exist; some archives don't include an explicit entry
			// for parent directories
			parentDir := filepath.Dir(outputPath)
			if err = os.MkdirAll(parentDir, 0o755); err != nil {
				return fmt.Errorf("create parent dirs for %s failed: %w", f.NameInArchive, err)
			}

			var extractedFile *os.File
			if extractedFile, err = os.OpenFile(outputPath, openFlags, f.Mode()); err != nil {
				return fmt.Errorf("create file failed: %w", err)
			}

			reader, err := f.Open()
			if err != nil {
				return fmt.Errorf("archive content open failed: %w", err)
			}
			defer reader.Close()

			if _, err = io.Copy(extractedFile, reader); err != nil {
				if closeErr := extractedFile.Close(); closeErr != nil {
					return fmt.Errorf("copy failed: %w; close also failed: %v", err, closeErr)
				}
				return fmt.Errorf("copy failed: %w", err)
			}
			if err = extractedFile.Close(); err != nil {
				return fmt.Errorf("close failed: %w", err)
			}
		}
		return nil
	})
}
