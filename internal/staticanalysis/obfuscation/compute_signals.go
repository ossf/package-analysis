package obfuscation

import (
	"math"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/detections"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stats"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stringentropy"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
	"github.com/ossf/package-analysis/internal/utils"
)

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
ComputeSignals creates a FileSignals object based on the data obtained from ParseSingle
for a given file. These signals may be useful to determine whether the code is obfuscated.
*/
func ComputeSignals(rawData parsing.Data) FileSignals {
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
