package obfuscation

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/utils"
)

// Result holds all data produced by obfuscation analysis (see Analyze() in analyze.go).
type Result struct {
	// ExcludedFiles is a list of package files that were excluded from analysis,
	// e.g. because they could not be parsed by any supported parser.
	ExcludedFiles []string `json:"excluded_files"`

	// Signals contains an obfuscation.FileSignals object that is useful for detecting suspicious files.
	Signals []FileSignals `json:"signals"`
}

func (r Result) String() string {
	signalsStrings := utils.Transform(r.Signals, func(s FileSignals) string { return s.String() })

	parts := []string{
		fmt.Sprintf("excluded files:\n%s", strings.Join(r.ExcludedFiles, ", ")),
		fmt.Sprintf("file signals:\n%s", strings.Join(signalsStrings, "\n==================\n")),
	}

	return strings.Join(parts, "\n\n-----------------------------\n\n")
}
