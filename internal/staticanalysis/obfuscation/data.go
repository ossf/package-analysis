package obfuscation

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stats"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

type RawData struct { // TODO rename to FileData
	LineLengths    map[int]int
	Identifiers    []token.Identifier
	StringLiterals []token.String
	IntLiterals    []token.Int
	FloatLiterals  []token.Float
	Comments       []token.Comment
}

type Signals struct { // TODO rename to FileSignals
	// StringLengths is a map from length (in characters) to number of
	// string literals in the file having that length. If a length key is
	// missing, it is assumed to be zero.
	StringLengths map[int]int

	// StringEntropySummary provides sample statistics for the set of entropy
	// values calculated on each string literal. Character probabilities for the
	// entropy calculation are estimated empirically from aggregated counts
	// of characters across all string literals in the file.
	StringEntropySummary stats.SampleStatistics

	// CombinedStringEntropy is the entropy of the string obtained from
	// concatenating all string literals in the file together. It may be used
	// to normalise the values in StringEntropySummary
	CombinedStringEntropy float64

	// IdentifierLengths is a map from length (in characters) to number of
	// identifiers in the file having that length. If a length key is missing,
	// it is assumed to be zero.
	IdentifierLengths map[int]int

	// IdentifierEntropySummary provides sample statistics for the set of entropy
	// values calculated on each identifier. Character probabilities for the
	// entropy calculation are estimated empirically from aggregated counts
	// of characters across all identifiers in the file.
	IdentifierEntropySummary stats.SampleStatistics

	// CombinedIdentifierEntropy is the entropy of the string obtained from
	// concatenating all identifiers in the file together. It may be used to
	// normalise the values in IdentifierEntropySummary
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
		fmt.Sprintf("line lengths\n%v\n", rd.LineLengths),
		fmt.Sprintf("identifiers\n%v\n", rd.Identifiers),
		fmt.Sprintf("string literals\n%v\n", rd.StringLiterals),
		fmt.Sprintf("integer literals\n%v\n", rd.IntLiterals),
		fmt.Sprintf("float literals\n%v\n", rd.FloatLiterals),
	}
	return strings.Join(parts, "\n-------------------\n")
}

func (s Signals) String() string {
	parts := []string{
		fmt.Sprintf("string lengths: %v", s.StringLengths),
		fmt.Sprintf("string entropy: %s", s.StringEntropySummary),
		fmt.Sprintf("combined string entropy: %f", s.CombinedStringEntropy),

		fmt.Sprintf("identifier lengths: %v", s.IdentifierLengths),
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
