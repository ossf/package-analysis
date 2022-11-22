package detection

import (
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stats"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stringentropy"
)

// characterAnalysis performs analysis on a collection of string symbols, returning:
// - Stats summary of symbol (string) lengths
// - Stats summary of symbol (string) entropies
// - Entropy of all symbols concatenated together
func characterAnalysis(symbols []string) (
	lengthSummary stats.SampleStatistics,
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

	lengthSummary = stats.Summarise(lengths)
	entropySummary = stats.Summarise(entropies)
	combinedEntropy = stringentropy.CalculateEntropy(strings.Join(symbols, ""), nil)
	return
}

/*
ComputeSignals operates on the data obtained from CollectData, and produces
signals which may be useful to determine whether the code is obfuscated.

Current signals:
  - Statistics of string literal lengths and string entropies
  - Statistics of identifier lengths and string entropies

TODO Planned signals
  - analysis of numeric literal arrays (entropy)
*/
func ComputeSignals(rawData obfuscation.RawData) obfuscation.Signals {
	signals := obfuscation.Signals{}
	signals.StringLengthSummary, signals.StringEntropySummary, signals.CombinedStringEntropy =
		characterAnalysis(rawData.StringLiterals)

	signals.IdentifierLengthSummary, signals.IdentifierEntropySummary, signals.CombinedIdentifierEntropy =
		characterAnalysis(rawData.Identifiers)

	return signals
}
