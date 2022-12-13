package obfuscation

import (
	"math"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stats"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stringentropy"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
	"github.com/ossf/package-analysis/internal/utils"
)

// characterAnalysis performs analysis on a collection of string symbols, returning:
// - Stats summary of symbol (string) lengths
// - Stats summary of symbol (string) entropies
// - Entropy of all symbols concatenated together
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
ComputeSignals operates on the data obtained from CollectData, and produces
signals which may be useful to determine whether the code is obfuscated.

Current signals:
  - Statistics of string literal lengths and string entropies
  - Statistics of identifier lengths and string entropies

TODO Planned signals
  - analysis of numeric literal arrays (entropy)
*/
func ComputeSignals(rawData RawData) Signals {
	signals := Signals{}

	literals := utils.Transform(rawData.StringLiterals, func(s token.String) string { return s.Value })
	signals.StringLengths, signals.StringEntropySummary, signals.CombinedStringEntropy =
		characterAnalysis(literals)

	identifierNames := utils.Transform(rawData.Identifiers, func(i token.Identifier) string { return i.Name })
	signals.IdentifierLengths, signals.IdentifierEntropySummary, signals.CombinedIdentifierEntropy =
		characterAnalysis(identifierNames)

	return signals
}

func NoSignals() Signals {
	return Signals{
		StringLengths:             map[int]int{},
		StringEntropySummary:      stats.NoData(),
		CombinedStringEntropy:     math.NaN(),
		IdentifierLengths:         map[int]int{},
		IdentifierEntropySummary:  stats.NoData(),
		CombinedIdentifierEntropy: math.NaN(),
	}
}

// RemoveNaNs replaces all NaN values in this object with zeros
func RemoveNaNs(s Signals) Signals {
	replaced := Signals{
		StringLengths:             s.StringLengths,
		StringEntropySummary:      s.StringEntropySummary.ReplaceNaNs(0),
		CombinedStringEntropy:     s.CombinedStringEntropy,
		IdentifierLengths:         s.IdentifierLengths,
		IdentifierEntropySummary:  s.IdentifierEntropySummary.ReplaceNaNs(0),
		CombinedIdentifierEntropy: s.CombinedIdentifierEntropy,
	}

	if math.IsNaN(replaced.CombinedStringEntropy) {
		replaced.CombinedStringEntropy = 0.0
	}
	if math.IsNaN(replaced.CombinedIdentifierEntropy) {
		replaced.CombinedIdentifierEntropy = 0.0
	}

	return replaced
}
