package staticanalysis

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/staticanalysis/linelengths"
	"github.com/ossf/package-analysis/internal/utils"
)

// BasicPackageData records basic information about a package
type BasicPackageData struct {
	Files map[string]BasicFileData `json:"files"`
}

// BasicFileData records various information about a file that can be determined
// without parsing it using a programming language parser.
type BasicFileData struct {
	// FileType records the output of the `file` command run on that file
	FileType string

	// Size records the size of the file (as reported by the filesystem)
	Size int64

	// Hash records the SHA256sum hash of the file
	Hash string

	// LineLengthCounts records the count of lines of each length in the file,
	// where a line is defined as all characters up to a newline.
	LineLengthCounts map[int]int
}

func (bd BasicFileData) String() string {
	// print line length counts in ascending order
	lineLengths := maps.Keys(bd.LineLengthCounts)
	slices.Sort(lineLengths)
	lineLengthStrings := make([]string, len(bd.LineLengthCounts))
	for length := range lineLengths {
		count := bd.LineLengthCounts[length]
		lineLengthStrings = append(lineLengthStrings, fmt.Sprintf("length = %4d, count = %2d", length, count))
	}

	parts := []string{
		fmt.Sprintf("file type: %v\n", bd.FileType),
		fmt.Sprintf("size: %v\n", bd.Size),
		fmt.Sprintf("hash: %v\n", bd.Hash),
		fmt.Sprintf("line lengths:\n%s", strings.Join(lineLengthStrings, "\n")),
	}
	return strings.Join(parts, "\n")
}

// GetBasicFileData collects basic file information for the file at the given path
// Errors are logged rather than returned, since some operations may succeed even if others fail
func GetBasicFileData(path, pathInArchive string) BasicFileData {
	result := BasicFileData{}

	// file size
	if fileInfo, err := os.Stat(path); err != nil {
		result.Size = -1 // error value
		log.Error("Error stat file", "path", pathInArchive, "error", err)
	} else {
		result.Size = fileInfo.Size()
	}

	// file hash
	if hash, err := utils.HashFile(path); err != nil {
		log.Error("Error hashing file", "path", pathInArchive, "error", err)
	} else {
		result.Hash = hash
	}

	// file type
	cmd := exec.Command("file", "--brief", path)
	if fileCmdOutput, err := cmd.Output(); err != nil {
		log.Error("Error running file command", "path", pathInArchive, "error", err)
	} else {
		result.FileType = strings.TrimSpace(string(fileCmdOutput))
	}

	// line lengths o
	lineLengths, err := linelengths.GetLineLengths(path, "")
	if err != nil {
		log.Error("Error collecting line lengths", "path", pathInArchive, "error", err)
	} else {
		result.LineLengthCounts = lineLengths
	}

	return result
}
