package obfuscation

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stats"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
)

type EscapedString struct {
	RawLiteral       string
	LevenshteinRatio float64
}

type FileSignals struct {
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

	// SuspiciousIdentifiers holds lists of identifiers that are deemed 'suspicious'
	// (i.e. indicative of obfuscation) according to certain rules. The keys of the
	// map are the rule names, and the values are the identifiers matching each rule.
	// See
	SuspiciousIdentifiers map[string][]string

	// Base64Strings holds a list of (substrings of) string literals found in the
	// file that match a base64 regex pattern. This patten has a minimum matching
	// length in order to reduce the number of false positives.
	Base64Strings []string

	// HexStrings holds a list of (substrings of) string literals found in the
	// file that contain long (>8 digits) hexadecimal digit sequences.
	HexStrings []string

	// EscapedStrings contain string literals that contain large amount of escape
	// characters, which may indicate obfuscation
	EscapedStrings []EscapedString
}

// AnalysisResult holds all the information obtained from
// obfuscation analysis of a single package artifact.
type AnalysisResult struct {
	// FileData maps analysed file names to the FileData collected for that file.
	FileData map[string]parsing.Data

	// FileSignals maps file names in the package to
	// the FileSignals computed for that file.
	FileSignals map[string]FileSignals

	// PackageData contains aggregated information from
	// all files and/or signals that can only be computed
	// from global information about the package
	PackageData struct{}

	// ExcludedFiles is a list of package files that were
	// excluded from analysis, e.g. because all supported
	// parsers encountered syntax errors when analysing the file.
	ExcludedFiles []string

	// The maps below contain data on all files in the package,
	// regardless of whether they were included in analysis.

	// FileSizes maps file names to file sizes in bytes.
	FileSizes map[string]int64

	// FileHashes maps file names to SHA256 hashes of the files.
	FileHashes map[string]string

	// FileTypes maps file names to the output of the `file`
	// command run on that file
	FileTypes map[string]string
}

func (s FileSignals) String() string {
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

	for filename, rawData := range ar.FileData {
		fileRawDataStrings = append(fileRawDataStrings, fmt.Sprintf("== %s ==\n%s", filename, rawData))
	}
	for filename, signals := range ar.FileSignals {
		fileSignalsStrings = append(fileSignalsStrings, fmt.Sprintf("== %s ==\n%s\n", filename, signals))
	}

	parts := []string{
		fmt.Sprintf("File Raw Data\n%s", strings.Join(fileRawDataStrings, "\n\n")),
		fmt.Sprintf("File Signals\n%s", strings.Join(fileSignalsStrings, "\n\n")),
		fmt.Sprintf("Package Data\n%s", ar.PackageData),
		fmt.Sprintf("Excluded files\n%s", strings.Join(ar.ExcludedFiles, "\n")),
		fmt.Sprintf("File sizes\n%v", ar.FileSizes),
	}

	return strings.Join(parts, "\n\n########################\n\n")
}
