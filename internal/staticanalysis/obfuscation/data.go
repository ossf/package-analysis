package obfuscation

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stats"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

type RawData struct {
	Identifiers    []token.Identifier
	StringLiterals []token.String
	IntLiterals    []token.Int
	FloatLiterals  []token.Float
	Comments       []token.Comment
}

type Signals struct {
	StringLengthSummary       stats.SampleStatistics
	StringEntropySummary      stats.SampleStatistics
	CombinedStringEntropy     float64
	IdentifierLengthSummary   stats.SampleStatistics
	IdentifierEntropySummary  stats.SampleStatistics
	CombinedIdentifierEntropy float64
}

// AnalysisResult holds all the information obtained from
// obfuscation analysis of a single package artifact.
type AnalysisResult struct {
	// FileRawData maps file names in the package to
	// the RawData collected for that file.
	FileRawData map[string]RawData

	// FileSignals maps file names in the package to
	// the Signals computed for that file.
	FileSignals map[string]Signals

	// PackageSignals contains aggregated signals from
	// all files and/or signals that can only be computed
	// from global information about the package
	PackageSignals struct{}

	// ExcludedFiles is a list of package files that were
	// excluded from analysis, e.g. because all supported
	// parsers encountered syntax errors when analysing the file.
	ExcludedFiles []string
}

func (rd RawData) String() string {
	parts := []string{
		fmt.Sprintf("Identifiers\n%v", rd.Identifiers),
		fmt.Sprintf("String Literals\n%v", rd.StringLiterals),
		fmt.Sprintf("Integer Literals\n%v", rd.IntLiterals),
		fmt.Sprintf("Float Literals\n%v", rd.FloatLiterals),
	}
	return strings.Join(parts, "\n-------------------\n")
}

func (s Signals) String() string {
	parts := []string{
		fmt.Sprintf("string length: %s", s.StringLengthSummary),
		fmt.Sprintf("string entropy: %s", s.StringEntropySummary),
		fmt.Sprintf("combined string entropy: %f", s.CombinedStringEntropy),

		fmt.Sprintf("identifier length: %s", s.IdentifierLengthSummary),
		fmt.Sprintf("identifier entropy: %s", s.IdentifierEntropySummary),
		fmt.Sprintf("combined identifier entropy: %f", s.CombinedIdentifierEntropy),
	}
	return strings.Join(parts, "\n")
}

func (ar AnalysisResult) String() string {
	fileRawDataStrings := make([]string, 0)
	fileSignalsStrings := make([]string, 0)

	for filename, rawData := range ar.FileRawData {
		fileRawDataStrings = append(fileRawDataStrings, fmt.Sprintf("== %s ==\n%s", filename, rawData))
	}
	for filename, signals := range ar.FileSignals {
		fileSignalsStrings = append(fileSignalsStrings, fmt.Sprintf("== %s ==\n%s\n", filename, signals))
	}

	parts := []string{
		fmt.Sprintf("File Raw Data\n%s", strings.Join(fileRawDataStrings, "\n\n")),
		fmt.Sprintf("File Signals\n%s", strings.Join(fileSignalsStrings, "\n\n")),
		fmt.Sprintf("Package Signals\n%s", ar.PackageSignals),
		fmt.Sprintf("Excluded Files\n%s", strings.Join(ar.ExcludedFiles, "\n")),
	}

	return strings.Join(parts, "\n\n########################\n\n")
}
