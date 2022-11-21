package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/ossf/package-analysis/internal/log"
)

// makeFileHeader initialises a record for a directory entry in a tar file.
func makeDirHeader(name string) *tar.Header {
	return &tar.Header{
		Typeflag: tar.TypeDir,
		Name:     name,
		Mode:     0777,
		Uid:      os.Geteuid(),
		Gid:      os.Getegid(),
	}
}

// makeFileHeader initialises a record for a file entry in a tar file.
// size is constrained to fit in an int to allow easier writing of the file
func makeFileHeader(name string, size int) *tar.Header {
	return &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     name,
		Size:     int64(size),
		Mode:     0666,
		Uid:      os.Geteuid(),
		Gid:      os.Getegid(),
	}

}

func createTgzFile(name string, headers []*tar.Header) (path string, err error) {
	tgzFile, err := os.CreateTemp("", name)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}

	path = tgzFile.Name()

	defer func() {
		closeErr := tgzFile.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close tgz file: %v", closeErr)
		}
	}()

	gzWriter := gzip.NewWriter(tgzFile)

	defer func() {
		closeErr := gzWriter.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close gz writer: %v", err)
		}
	}()

	tarWriter := tar.NewWriter(gzWriter)

	for _, header := range headers {
		if err = tarWriter.WriteHeader(header); err != nil {
			return
		}
		size := int(header.Size) // constrained to int
		if size > 0 {
			// write # bytes to file
			bytes := make([]byte, size)
			for i := 0; i < size; i++ {
				bytes[i] = '\n'
			}
			l, writeErr := tarWriter.Write(bytes)
			if writeErr != nil {
				err = writeErr
				return
			}
			if l != size {
				err = fmt.Errorf("expected to write %d bytes but wrote %d bytes", size, l)
				return
			}
		}
	}

	err = tarWriter.Close()
	return
}

func doExtractionTest(testName string, testHeaders []*tar.Header, runChecks func(outputDir string) error) (err error) {
	testFile, err := createTgzFile(testName, testHeaders)

	if testFile == "" || err != nil {
		return fmt.Errorf("failed to create test tgz file: %v", err)
	}

	defer func() {
		if removeErr := os.Remove(testFile); removeErr != nil && err == nil {
			err = fmt.Errorf("failed to remove test tgz file: %v", removeErr)
		}
	}()

	extractDir, err := os.MkdirTemp("", testName)
	if err != nil {
		return fmt.Errorf("failed to create temp dir for extraction: %v", err)
	}

	defer func() {
		if removeErr := os.RemoveAll(extractDir); removeErr != nil && err == nil {
			err = fmt.Errorf("failed to remove extracted files dir: %v", removeErr)
		}
	}()

	log.Initalize("")
	err = ExtractTarGzFile(testFile, extractDir)
	if err != nil {
		return fmt.Errorf("extract failed: %v", err)
	}

	return runChecks(extractDir)
}

func TestExtractSimpleTarGzFile(t *testing.T) {
	testName := "package-analysis-test-tgz-simple"

	testHeaders := []*tar.Header{
		makeDirHeader("test"),
		makeFileHeader("test/1.txt", 10),
	}

	err := doExtractionTest(testName, testHeaders, func(extractDir string) error {
		dirInfo, err := os.Stat(path.Join(extractDir, "test"))
		if err != nil {
			return fmt.Errorf("stat extracted dir: %v", err)
		}
		if dirInfo.Name() != "test" {
			return fmt.Errorf("expected extracted directory name 'test', got %s", dirInfo.Name())
		}
		if !dirInfo.IsDir() {
			return fmt.Errorf("expected to extract directory but it was not a directory")
		}

		fileInfo, err := os.Stat(path.Join(extractDir, "test", "1.txt"))
		if err != nil {
			return fmt.Errorf("stat extracted file: %v", err)
		}
		if fileInfo.Name() != "1.txt" {
			return fmt.Errorf("expected to extract file with name '1.txt' but it has name %s", fileInfo.Name())
		}
		if fileInfo.Size() != 10 {
			return fmt.Errorf("expected to extract file with size 10 but it has size %d", fileInfo.Size())
		}
		if fileInfo.IsDir() {
			return fmt.Errorf("expected to extract file but it was a directory")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Error: %v", err)
	}
}

func TestExtractZipSlip(t *testing.T) {
	testName := "package-analysis-test-tgz-zipslip"

	testHeaders := []*tar.Header{
		makeDirHeader("test"),
		makeFileHeader("test/../../bad.txt", 1),
	}

	err := doExtractionTest(testName, testHeaders, func(outputDir string) error {
		t.Fatal("Extraction should have returned an error")
		return nil
	})

	if err == nil || !strings.Contains(err.Error(), "archive path escapes output dir") {
		t.Errorf("Error should be about path escaping output dir, instead got %v", err)
	}
}

func TestExtractAbsolutePathTarGzFile(t *testing.T) {
	testName := "package-analysis-test-tgz-abs-path"

	testHeaders := []*tar.Header{
		makeDirHeader("/test"),
		makeFileHeader("/test/2.txt", 0),
	}

	err := doExtractionTest(testName, testHeaders, func(extractDir string) error {
		dirInfo, err := os.Stat(path.Join(extractDir, "test"))
		if err != nil {
			return fmt.Errorf("stat extracted dir: %v", err)
		}
		if dirInfo.Name() != "test" {
			return fmt.Errorf("expected extracted directory name 'test', got %s", dirInfo.Name())
		}
		if !dirInfo.IsDir() {
			return fmt.Errorf("expected to extract directory but it was not a directory")
		}

		fileInfo, err := os.Stat(path.Join(extractDir, "test", "2.txt"))
		if err != nil {
			return fmt.Errorf("stat extracted file: %v", err)
		}
		if fileInfo.Name() != "2.txt" {
			return fmt.Errorf("expected to extract file with name '1.txt' but it has name %s", fileInfo.Name())
		}
		if fileInfo.Size() != 0 {
			return fmt.Errorf("expected to extract file with size 0 but it has size %d", fileInfo.Size())
		}
		if fileInfo.IsDir() {
			return fmt.Errorf("expected to extract file but it was a directory")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Error: %v", err)
	}
}
