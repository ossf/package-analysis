package obfuscation

import (
	"fmt"
	"math"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/detections"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stats"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stringentropy"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
	"github.com/ossf/package-analysis/internal/utils"
)

// FileSignals holds information related to the presence of obfuscated code in a single file
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

func (s FileSignals) String() string {
	parts := []string{
		fmt.Sprintf("string lengths: %v", s.StringLengths),
		fmt.Sprintf("string entropy: %s", s.StringEntropySummary),
		fmt.Sprintf("combined string entropy: %f", s.CombinedStringEntropy),

		fmt.Sprintf("identifier lengths: %v", s.IdentifierLengths),
		fmt.Sprintf("identifier entropy: %s", s.IdentifierEntropySummary),
		fmt.Sprintf("combined identifier entropy: %f", s.CombinedIdentifierEntropy),

		fmt.Sprintf("suspicious identifiers: %v", s.SuspiciousIdentifiers),
		fmt.Sprintf("potential base64 strings: %v", s.Base64Strings),
		fmt.Sprintf("hex strings: %v", s.HexStrings),
		fmt.Sprintf("escaped strings: %v", s.EscapedStrings),
	}
	return strings.Join(parts, "\n")
}

type EscapedString struct {
	RawLiteral       string
	LevenshteinRatio float64
}

/*
characterAnalysis performs analysis on a collection of string symbols, returning:
- Counts of symbol (string) lengths
- Stats summary of symbol (string) entropies
- Entropy of all symbols concatenated together
*/
func characterAnalysis(symbols []string) (
	lengthCounts map[int]int,
	entropySummary stats.SampleStatistics,
	combinedEntropy float64,
) {
	// measure character probabilities by looking at entire set of strings
	characterProbs := stringentropy.CharacterProbabilities(symbols)

	var entropies []float64
	var lengths []int
	for _, s := range symbols {
		entropies = append(entropies, stringentropy.CalculateEntropy(s, characterProbs))
		lengths = append(lengths, len(s))
	}

	lengthCounts = stats.CountDistinct(lengths)
	entropySummary = stats.Summarise(entropies)
	combinedEntropy = stringentropy.CalculateEntropy(strings.Join(symbols, ""), nil)
	return
}

/*
ComputeFileSignals creates a FileSignals object based on the data obtained from ParseFile
for a given file. These signals may be useful to determine whether the code is obfuscated.
*/
func ComputeFileSignals(rawData parsing.SingleResult) FileSignals {
	signals := FileSignals{}

	literals := utils.Transform(rawData.StringLiterals, func(s token.String) string { return s.Value })
	signals.StringLengths, signals.StringEntropySummary, signals.CombinedStringEntropy =
		characterAnalysis(literals)

	identifierNames := utils.Transform(rawData.Identifiers, func(i token.Identifier) string { return i.Name })
	signals.IdentifierLengths, signals.IdentifierEntropySummary, signals.CombinedIdentifierEntropy =
		characterAnalysis(identifierNames)

	signals.SuspiciousIdentifiers = map[string][]string{}
	for ruleName, pattern := range detections.SuspiciousIdentifierPatterns {
		signals.SuspiciousIdentifiers[ruleName] = []string{}
		for _, name := range identifierNames {
			if pattern.MatchString(name) {
				signals.SuspiciousIdentifiers[ruleName] = append(signals.SuspiciousIdentifiers[ruleName], name)
			}
		}
	}

	signals.Base64Strings = []string{}
	signals.HexStrings = []string{}
	signals.EscapedStrings = []EscapedString{}
	for _, s := range rawData.StringLiterals {
		signals.Base64Strings = append(signals.Base64Strings, detections.FindBase64Substrings(s.Value)...)
		signals.HexStrings = append(signals.HexStrings, detections.FindHexSubstrings(s.Value)...)
		if detections.IsHighlyEscaped(s, 8, 0.25) {
			escapedString := EscapedString{
				RawLiteral:       s.Raw,
				LevenshteinRatio: detections.LevenshteinRatio(s),
			}
			signals.EscapedStrings = append(signals.EscapedStrings, escapedString)
		}
	}

	return signals
}

func NoSignals() FileSignals {
	return FileSignals{
		StringLengths:             map[int]int{},
		StringEntropySummary:      stats.NoData(),
		CombinedStringEntropy:     math.NaN(),
		IdentifierLengths:         map[int]int{},
		IdentifierEntropySummary:  stats.NoData(),
		CombinedIdentifierEntropy: math.NaN(),
	}
}

// RemoveNaNs replaces all NaN values in this object with zeros
func RemoveNaNs(s *FileSignals) {
	s.StringEntropySummary = s.StringEntropySummary.ReplaceNaNs(0)
	s.IdentifierEntropySummary = s.IdentifierEntropySummary.ReplaceNaNs(0)

	if math.IsNaN(s.CombinedStringEntropy) {
		s.CombinedStringEntropy = 0.0
	}
	if math.IsNaN(s.CombinedIdentifierEntropy) {
		s.CombinedIdentifierEntropy = 0.0
	}
}
