package staticanalysis

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/staticanalysis/externalcmd"
	"github.com/ossf/package-analysis/internal/staticanalysis/linelengths"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/internal/utils/valuecounts"
)

// BasicPackageData records basic information about files in a package,
// mapping file path within the archive to BasicFileData about that file.
type BasicPackageData struct {
	Files []BasicFileData `json:"files"`
}

// BasicFileData records various information about a file that can be determined
// without parsing it using a programming language parser.
type BasicFileData struct {
	// Filename records the path to the file within the package archive
	Filename string `json:"filename"`

	// Description records the output of the `file` command run on that file.
	Description string `json:"filetype"`

	// Size records the size of the file (as reported by the filesystem).
	Size int64 `json:"size"`

	// Hash records the SHA256sum hash of the file.
	Hash string `json:"hash"`

	// LineLengths records the counts of line lengths in the file,
	// where a line is defined as all characters up to a newline.
	LineLengths valuecounts.ValueCounts `json:"line_lengths"`
}

func (bd BasicFileData) String() string {
	parts := []string{
		fmt.Sprintf("file type: %v\n", bd.Description),
		fmt.Sprintf("size: %v\n", bd.Size),
		fmt.Sprintf("hash: %v\n", bd.Hash),
		fmt.Sprintf("line lengths: %v\n", bd.LineLengths),
	}
	return strings.Join(parts, "\n")
}

// fileCmdInputArgs describes how to pass file arguments to the `file` command.
type fileCmdArgsHandler struct{}

func (h fileCmdArgsHandler) SingleFileArg(filePath string) []string {
	return []string{filePath}
}

func (h fileCmdArgsHandler) FileListArg(fileListPath string) []string {
	return []string{"--files-from", fileListPath}
}

func (h fileCmdArgsHandler) ReadStdinArg() []string {
	// reads file list from standard input
	return h.FileListArg("-")
}

func getFileDescriptions(fileList []string) ([]string, error) {
	workingDir, err := os.MkdirTemp("", "package-analysis-basic-data-*")
	if err != nil {
		return nil, fmt.Errorf("error creating temp file for file type analysis: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(workingDir); err != nil {
			log.Error("could not remove working directory", "path", workingDir, "error", err)
		}
	}()

	cmd := exec.Command("file", "--brief")
	input := externalcmd.MultipleFileInput(fileList)

	if err := input.SendTo(cmd, fileCmdArgsHandler{}, workingDir); err != nil {
		return nil, fmt.Errorf("failed to prepare input for file type analysis: %w", err)
	}

	fileCmdOutput, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error running file command: %w", err)
	}

	fileTypesString := strings.TrimSpace(string(fileCmdOutput))
	if fileTypesString == "" {
		// no files input, probably
		return []string{}, nil
	}

	// command output is newline-separated list of file types,
	// with the order matching the input file list.
	return strings.Split(fileTypesString, "\n"), nil
}

/*
GetBasicData collects basic file information for the specified files
Errors are logged rather than returned, since failures in analysing
some files should not prevent the analysis of other files.

pathInArchive maps the absolute paths in fileList to relative paths
in the package archive, to use for results.
*/
func GetBasicData(fileList []string, pathInArchive func(absolutePath string) string) (*BasicPackageData, error) {
	// First, run file in batch processing mode to get all the file types at once.
	// Then, file size, hash and line lengths can be done in a simple loop

	fileDescriptions, err := getFileDescriptions(fileList)
	if err != nil {
		return nil, err
	}
	if len(fileDescriptions) != len(fileList) {
		return nil, fmt.Errorf("file type analysis returned mismatched number of results")
	}

	result := BasicPackageData{
		Files: []BasicFileData{},
	}

	for index, filePath := range fileList {
		archivePath := pathInArchive(filePath)
		description := fileDescriptions[index]

		var fileSize int64
		if fileInfo, err := os.Stat(filePath); err != nil {
			fileSize = -1 // error value
			log.Error("Error during stat file", "path", archivePath, "error", err)
		} else {
			fileSize = fileInfo.Size()
		}

		var fileHash string
		if hash, err := utils.SHA256Hash(filePath); err != nil {
			log.Error("Error hashing file", "path", archivePath, "error", err)
		} else {
			// append label to hash
			fileHash = "sha256:" + hash
		}

		var lineLengths valuecounts.ValueCounts
		if ll, err := linelengths.GetLineLengths(filePath, ""); err != nil {
			log.Error("Error counting line lengths", "path", archivePath, "error", err)
		} else {
			lineLengths = valuecounts.Count(ll)
		}

		result.Files = append(result.Files, BasicFileData{
			Filename:    archivePath,
			Description: description,
			Size:        fileSize,
			Hash:        fileHash,
			LineLengths: lineLengths,
		})
	}

	return &result, nil
}
