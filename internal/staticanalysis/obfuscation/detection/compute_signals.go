package detection

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stats"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stringentropy"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing/js"
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
CollectData parses the given input (source file or raw source string) and records data which can be
later processed into signals that which may be useful to determine whether the code is obfuscated.
The input is assumed to be valid JavaScript source. If jsSourceFile is empty, the string will be parsed.

In Javascript, there is little distinction between integer and floating point literals - they are
all parsed as floating point. This function will record a numeric literal as an integer if it can be
converted to an integer using strconv.Atoi(), otherwise it will be recorded as a floating point literal.

Current data collected:
  - list of identifiers (e.g. variable, function, and class names, loop labels)
  - lists of string, integer and floating point literals

TODO planned data
  - recording of arrays of either string literals or numeric data
*/
func CollectData(jsParserPath, jsSourceFile string, jsSourceString string, printDebug bool) (*RawData, error) {
	parseResult, rawJson, err := js.ParseJS(jsParserPath, jsSourceFile, jsSourceString)
	if printDebug {
		println("\nRaw JSON:\n", rawJson)
	}
	if err != nil && parseResult == nil {
		fmt.Printf("Error occured while reading %s: %v\n", jsSourceFile, err)
		return nil, err
	}

	data := RawData{}

	for _, d := range parseResult.Literals {
		if d.GoType == "string" {
			data.StringLiterals = append(data.StringLiterals, d.Value.(string))
		} else if d.GoType == "float64" {
			// check if it can be an int
			if intValue, err := strconv.Atoi(d.RawValue); err == nil {
				data.IntLiterals = append(data.IntLiterals, intValue)
			} else {
				data.FloatLiterals = append(data.FloatLiterals, d.Value.(float64))
			}
		}
	}

	for _, ident := range parseResult.Identifiers {
		switch ident.Type {
		case parsing.Function:
			fallthrough
		case parsing.Class:
			fallthrough
		case parsing.Parameter:
			fallthrough
		case parsing.Property:
			fallthrough
		case parsing.Variable:
			data.Identifiers = append(data.Identifiers, ident.Name)
		}
	}

	return &data, nil
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
	signals.StringLengthSummary, signals.StringEntropySummary, signals.CombinedStringEntropy =
		characterAnalysis(rawData.StringLiterals)

	signals.IdentifierLengthSummary, signals.IdentifierEntropySummary, signals.CombinedIdentifierEntropy =
		characterAnalysis(rawData.Identifiers)

	return signals
}
