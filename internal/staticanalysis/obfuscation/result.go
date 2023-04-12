package obfuscation

import (
	"fmt"
	"strings"
)

// Result holds all data produced by obfuscation analysis (see Analyze() in analyze.go).
type Result struct {
	// ExcludedFiles is a list of package files that were excluded from analysis,
	// e.g. because they could not be parsed by any supported parser.
	ExcludedFiles []string `json:"excluded_files"`

	// Signals maps package file names to corresponding obfuscation.FileSignals
	// that are used to detect suspicious files.
	Signals map[string]FileSignals `json:"signals"`
}

func (r Result) String() string {
	fileSignalsStrings := make([]string, 0)

	for filename, signals := range r.Signals {
		fileSignalsStrings = append(fileSignalsStrings, fmt.Sprintf("== %s ==\n%s\n", filename, signals))
	}

	parts := []string{
		fmt.Sprintf("excluded files:\n%v", r.ExcludedFiles),
		fmt.Sprintf("file signals\n%s", strings.Join(fileSignalsStrings, "\n\n")),
	}

	return strings.Join(parts, "\n\n-----------------------------\n\n")
}
