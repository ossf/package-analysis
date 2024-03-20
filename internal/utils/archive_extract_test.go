package utils

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ossf/package-analysis/internal/log"
)

// makeFileHeader initialises a record for a directory entry in a tar file.
func makeDirHeader(name string) *tar.Header {
	return &tar.Header{
		Typeflag: tar.TypeDir,
		Name:     name,
		Mode:     0o777,
		Uid:      os.Geteuid(),
		Gid:      os.Getegid(),
	}
}

// makeFileHeader initialises a record for a file entry in a tar file.
//
// size is constrained to fit in an int to allow easier writing of the file.
func makeFileHeader(name string, size int) *tar.Header {
	return &tar.Header{
		Typeflag: tar.TypeReg,
		Name:     name,
		Size:     int64(size),
		Mode:     0o666,
		Uid:      os.Geteuid(),
		Gid:      os.Getegid(),
	}
}

func createTgzFile(path string, headers []*tar.Header) (err error) {
	tgzFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create temp archive file: %w", err)
	}

	path = tgzFile.Name()

	defer func() {
		closeErr := tgzFile.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close temp archive file: %w", closeErr)
		}
	}()

	gzWriter := gzip.NewWriter(tgzFile)

	defer func() {
		closeErr := gzWriter.Close()
		if closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close gzip writer: %w", err)
		}
	}()

	tarWriter := tar.NewWriter(gzWriter)

	for _, header := range headers {
		if err = tarWriter.WriteHeader(header); err != nil {
			return err
		}
		size := int(header.Size) // constrained to int
		if size > 0 {
			// write # bytes to file
			bytes := make([]byte, size)
			for i := 0; i < size; i++ {
				bytes[i] = '\n'
			}
			n, writeErr := tarWriter.Write(bytes)
			if writeErr != nil {
				return writeErr
			}
			if n != size {
				return fmt.Errorf("expected to write %d bytes but wrote %d bytes", size, n)
			}
		}
	}

	return tarWriter.Close()
}

func makePaths(t *testing.T, testName string) (workDir, archivePath, extractPath string, err error) {
	t.Helper()
	workDir = t.TempDir()
	archivePath = filepath.Join(workDir, testName+".tar.gz")
	extractPath = filepath.Join(workDir, "extracted")

	if err = os.Mkdir(extractPath, 0o700); err != nil {
		t.Fatalf("failed to create dir for extraction: %v", err)
	}

	return
}

func doExtractionTest(archivePath, extractPath string, archiveHeaders []*tar.Header, runChecks func() error) (err error) {
	if err = createTgzFile(archivePath, archiveHeaders); err != nil {
		return fmt.Errorf("failed to create test tgz file: %w", err)
	}

	log.Initialize("")

	if err = ExtractArchiveFile(archivePath, extractPath); err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}

	return runChecks()
}

func TestExtractSimpleTarGzFile(t *testing.T) {
	testName := "simple"

	_, archivePath, extractPath, err := makePaths(t, testName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	testHeaders := []*tar.Header{
		makeDirHeader("test"),
		makeFileHeader("test/1.txt", 10),
	}

	err = doExtractionTest(archivePath, extractPath, testHeaders, func() error {
		dirInfo, err := os.Stat(filepath.Join(extractPath, "test"))
		if err != nil {
			return fmt.Errorf("stat extracted dir: %w", err)
		}
		if dirInfo.Name() != "test" {
			return fmt.Errorf("expected extracted directory name 'test', got %s", dirInfo.Name())
		}
		if !dirInfo.IsDir() {
			return fmt.Errorf("expected to extract directory but it was not a directory")
		}

		fileInfo, err := os.Stat(filepath.Join(extractPath, "test", "1.txt"))
		if err != nil {
			return fmt.Errorf("stat extracted file: %w", err)
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

func TestExtractMissingParentDir(t *testing.T) {
	testName := "simple"

	_, archivePath, extractPath, err := makePaths(t, testName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	testHeaders := []*tar.Header{
		makeFileHeader("test/1.txt", 10),
	}

	err = doExtractionTest(archivePath, extractPath, testHeaders, func() error {
		dirInfo, err := os.Stat(filepath.Join(extractPath, "test"))
		if err != nil {
			return fmt.Errorf("stat extracted dir: %w", err)
		}
		if dirInfo.Name() != "test" {
			return fmt.Errorf("expected extracted directory name 'test', got %s", dirInfo.Name())
		}
		if !dirInfo.IsDir() {
			return fmt.Errorf("expected to extract directory but it was not a directory")
		}

		fileInfo, err := os.Stat(filepath.Join(extractPath, "test", "1.txt"))
		if err != nil {
			return fmt.Errorf("stat extracted file: %w", err)
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

func TestExtractAbsolutePathTarGzFile(t *testing.T) {
	testName := "abs-path"

	_, archivePath, extractPath, err := makePaths(t, testName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	testHeaders := []*tar.Header{
		makeDirHeader("/test"),
		makeFileHeader("/2.txt", 0),
	}

	err = doExtractionTest(archivePath, extractPath, testHeaders, func() error {
		dirInfo, err := os.Stat(filepath.Join(extractPath, "test"))
		if err != nil {
			return fmt.Errorf("stat extracted dir: %w", err)
		}
		if dirInfo.Name() != "test" {
			return fmt.Errorf("expected extracted directory name 'test', got %s", dirInfo.Name())
		}
		if !dirInfo.IsDir() {
			return fmt.Errorf("expected to extract directory but it was not a directory")
		}

		fileInfo, err := os.Stat(filepath.Join(extractPath, "2.txt"))
		if err != nil {
			return fmt.Errorf("stat extracted file: %w", err)
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

func TestExtractZipSlip(t *testing.T) {
	testName := "zipslip"

	_, archivePath, extractPath, err := makePaths(t, testName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	testHeaders := []*tar.Header{
		makeDirHeader("test"),
		makeFileHeader("test/../../bad.txt", 1),
	}

	err = doExtractionTest(archivePath, extractPath, testHeaders, func() error {
		t.Fatal("Extraction should have returned an error")
		return nil
	})

	if err == nil || !strings.Contains(err.Error(), "archive path escapes output dir") {
		t.Errorf("Error should be about path escaping output dir, instead got %v", err)
	}
}

func TestExtractZipSlip2(t *testing.T) {
	testName := "zipslip2"

	_, archivePath, extractPath, err := makePaths(t, testName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	// try and force writing into a similarly named directory
	similarlyNamedDir := extractPath + "FOO"
	err = os.Mkdir(similarlyNamedDir, 0o700)

	testHeaders := []*tar.Header{
		makeFileHeader(filepath.Join("..", filepath.Base(similarlyNamedDir), "bad2.txt"), 1),
	}

	err = doExtractionTest(archivePath, extractPath, testHeaders, func() error {
		bad2Info, err := os.Stat(filepath.Join(similarlyNamedDir, "bad2.txt"))
		if err == nil && bad2Info.Size() == 1 {
			t.Errorf("Found file in similarly named directory")
		}
		t.Fatal("Extraction should have returned an error")
		return nil
	})

	if err == nil || !strings.Contains(err.Error(), "archive path escapes output dir") {
		t.Errorf("Error should be about path escaping output dir, instead got %v", err)
	}
}

func TestExtractZipSlip3(t *testing.T) {
	testName := "zipslip3"

	workDir, archivePath, extractPath, err := makePaths(t, testName)
	if err != nil {
		t.Errorf("%v", err)
		return
	}

	testHeaders := []*tar.Header{
		makeFileHeader("../bad3.txt", 1),
	}

	err = doExtractionTest(archivePath, extractPath, testHeaders, func() error {
		bad3Info, err := os.Stat(filepath.Join(workDir, "bad3.txt"))
		if err == nil && bad3Info.Size() == 1 {
			t.Errorf("Found file in parent directory")
		}
		t.Fatal("Extraction should have returned an error")
		return nil
	})

	if err == nil || !strings.Contains(err.Error(), "archive path escapes output dir") {
		t.Errorf("Error should be about path escaping output dir, instead got %v", err)
	}
}
