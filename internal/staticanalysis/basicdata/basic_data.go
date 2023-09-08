package basicdata

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/linelengths"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/internal/utils/valuecounts"
)

// PackageData records basic information about files in a package,
// mapping file path within the archive to FileData about that file.
type PackageData struct {
	Files []FileData `json:"files"`
}

// FileData records various information about a file that can be determined
// without parsing it using a programming language parser.
type FileData struct {
	// Filename records the path to the file within the package archive
	Filename string `json:"filename"`

	// Description records the output of the `file` command run on that file.
	Description string `json:"description"`

	// Size records the size of the file (as reported by the filesystem).
	Size int64 `json:"size"`

	// SHA256 records the SHA256 hashsum of the file.
	SHA256 string `json:"sha256"`

	// LineLengths records the counts of line lengths in the file,
	// where a line is defined as all characters up to a newline.
	LineLengths valuecounts.ValueCounts `json:"line_lengths"`
}

func (bd FileData) String() string {
	parts := []string{
		fmt.Sprintf("filename: %v\n", bd.Filename),
		fmt.Sprintf("description: %v\n", bd.Description),
		fmt.Sprintf("size: %v\n", bd.Size),
		fmt.Sprintf("sha256: %v\n", bd.SHA256),
		fmt.Sprintf("line lengths: %v\n", bd.LineLengths),
	}
	return strings.Join(parts, "\n")
}

/*
Analyze collects basic file information for the specified files. Errors are logged
rather than returned where possible, to maximise the amount of data collected.
pathInArchive should return the relative path in the package archive, given an absolute
path to a file in the package. The relative path is used for the result data.
*/
func Analyze(ctx context.Context, paths []string, pathInArchive func(absolutePath string) string) (*PackageData, error) {
	if len(paths) == 0 {
		return &PackageData{Files: []FileData{}}, nil
	}

	descriptions, err := describeFiles(ctx, paths)
	haveDescriptions := true
	if err != nil {
		slog.ErrorContext(ctx, "failed to get file descriptions", "error", err)
		haveDescriptions = false
	}
	if len(descriptions) != len(paths) {
		slog.ErrorContext(ctx, fmt.Sprintf("describeFiles() returned %d results, expecting %d", len(descriptions), len(paths)))
		haveDescriptions = false
	}

	result := PackageData{
		Files: []FileData{},
	}

	for index, filePath := range paths {
		archivePath := pathInArchive(filePath)
		description := ""
		if haveDescriptions {
			description = descriptions[index]
		}

		var fileSize int64
		if fileInfo, err := os.Stat(filePath); err != nil {
			fileSize = -1 // error value
			slog.ErrorContext(ctx, "Error during stat file", "path", archivePath, "error", err)
		} else {
			fileSize = fileInfo.Size()
		}

		var sha265Sum string
		if hash, err := utils.SHA256Hash(filePath); err != nil {
			slog.ErrorContext(ctx, "Error hashing file", "path", archivePath, "error", err)
		} else {
			sha265Sum = hash
		}

		var lineLengths valuecounts.ValueCounts
		if ll, err := linelengths.GetLineLengths(filePath, ""); err != nil {
			slog.ErrorContext(ctx, "Error counting line lengths", "path", archivePath, "error", err)
		} else {
			lineLengths = valuecounts.Count(ll)
		}

		result.Files = append(result.Files, FileData{
			Filename:    archivePath,
			Description: description,
			Size:        fileSize,
			SHA256:      sha265Sum,
			LineLengths: lineLengths,
		})
	}

	return &result, nil
}
