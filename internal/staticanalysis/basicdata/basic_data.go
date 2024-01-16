package basicdata

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/linelengths"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/pkg/valuecounts"
)

// FileData records various information about a file that can be determined
// without parsing it using a programming language parser.
type FileData struct {
	// DetectedType records the output of the `file` command run on that file.
	DetectedType string

	// Size records the size of the file (as reported by the filesystem).
	Size int64

	// SHA256 records the SHA256 hashsum of the file.
	SHA256 string

	// LineLengths records the counts of line lengths in the file,
	// where a line is defined as all characters up to a newline.
	LineLengths valuecounts.ValueCounts
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

// Option allows controlling the behaviour of Analyze with non-required arguments.
type Option interface{ set(*analyzeConfig) }

// option implements Option.
type option func(*analyzeConfig)

func (o option) set(config *analyzeConfig) { o(config) }

// analyzeConfig stores all behaviour configuration for Analyze which is adjustable by Option.
type analyzeConfig struct {
	// withLineLengths enables line length analysis
	withLineLengths bool
	// formatPathFunc allows providing a custom transformation for file paths
	// when logging errors. For example, removing a common path prefix.
	formatPathFunc func(absPath string) string
}

func getDefaultAnalyzeConfig() analyzeConfig {
	return analyzeConfig{
		withLineLengths: true,
		formatPathFunc:  func(absPath string) string { return absPath },
	}
}

// SkipLineLengths disables collecting line length information during analysis, which is
// useful when the input files are known not to be text files (e.g. a package tarball).
func SkipLineLengths() Option {
	return option(func(config *analyzeConfig) {
		config.withLineLengths = false
	})
}

// FormatPaths uses the given function to transform absolute file paths
// before they are passed to logging.
func FormatPaths(formatPathFunc func(absPath string) string) Option {
	return option(func(config *analyzeConfig) {
		config.formatPathFunc = formatPathFunc
	})
}

/*
Analyze collects basic file information for the specified files. Errors are logged
rather than returned where possible, to maximise the amount of data collected.
Pass instances of Option to control which information is collected.
*/
func Analyze(ctx context.Context, paths []string, options ...Option) ([]FileData, error) {
	if len(paths) == 0 {
		return []FileData{}, nil
	}

	config := getDefaultAnalyzeConfig()
	for _, o := range options {
		o.set(&config)
	}

	var detectedTypes []string
	var haveDetectedTypes bool
	types, err := detectFileTypes(ctx, paths)
	haveDetectedTypes = true
	if err != nil {
		slog.ErrorContext(ctx, "failed to run file type detection", "error", err)
		haveDetectedTypes = false
	}
	if len(types) != len(paths) {
		slog.ErrorContext(ctx, fmt.Sprintf("detectFileTypes() returned %d results, expecting %d", len(detectedTypes), len(paths)))
		haveDetectedTypes = false
	}
	detectedTypes = types

	result := make([]FileData, len(paths))

	for index, filePath := range paths {
		formattedPath := config.formatPathFunc(filePath)
		detectedType := ""
		if haveDetectedTypes {
			detectedType = detectedTypes[index]
		}

		var fileSize int64
		if fileInfo, err := os.Stat(filePath); err != nil {
			fileSize = -1 // error value
			slog.ErrorContext(ctx, "Error during stat file", "file", formattedPath, "error", err)
		} else {
			fileSize = fileInfo.Size()
		}

		var sha265Sum string
		if hash, err := utils.SHA256Hash(filePath); err != nil {
			slog.ErrorContext(ctx, "Error hashing file", "file", formattedPath, "error", err)
		} else {
			sha265Sum = hash
		}

		var lineLengths valuecounts.ValueCounts
		if config.withLineLengths {
			if ll, err := linelengths.GetLineLengths(filePath, ""); err != nil {
				slog.ErrorContext(ctx, "Error counting line lengths", "file", formattedPath, "error", err)
			} else {
				lineLengths = valuecounts.Count(ll)
			}
		}

		result[index] = FileData{
			DetectedType: detectedType,
			Size:         fileSize,
			SHA256:       sha265Sum,
			LineLengths:  lineLengths,
		}
	}

	return result, nil
}
