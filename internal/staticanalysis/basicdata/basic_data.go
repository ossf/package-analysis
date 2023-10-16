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

// FileData records various information about a file that can be determined
// without parsing it using a programming language parser.
type FileData struct {
	// DetectedType records the output of the `file` command run on that file.
	DetectedType string `json:"detected_type"`

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
		fmt.Sprintf("detected type: %v\n", bd.DetectedType),
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
func Analyze(ctx context.Context, paths []string, pathInArchive func(absolutePath string) string) ([]FileData, error) {
	if len(paths) == 0 {
		return []FileData{}, nil
	}

	detectedTypes, err := detectFileTypes(ctx, paths)
	haveDetectedTypes := true
	if err != nil {
		slog.ErrorContext(ctx, "failed to run file type detection", "error", err)
		haveDetectedTypes = false
	}
	if len(detectedTypes) != len(paths) {
		slog.ErrorContext(ctx, fmt.Sprintf("detectFileTypes() returned %d results, expecting %d", len(detectedTypes), len(paths)))
		haveDetectedTypes = false
	}

	var result []FileData

	for index, filePath := range paths {
		archivePath := pathInArchive(filePath)
		detectedType := ""
		if haveDetectedTypes {
			detectedType = detectedTypes[index]
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

		result = append(result, FileData{
			DetectedType: detectedType,
			Size:         fileSize,
			SHA256:       sha265Sum,
			LineLengths:  lineLengths,
		})
	}

	return result, nil
}
