package obfuscation

import (
	"os"
	"strings"
	"testing"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stats"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stringentropy"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing/js"
	"github.com/ossf/package-analysis/internal/utils"
)

func symbolEntropySummary(symbols []string) stats.SampleStatistics {
	probs := stringentropy.CharacterProbabilities(symbols)
	entropies := utils.Transform(symbols, func(s string) float64 { return stringentropy.CalculateEntropy(s, probs) })
	return stats.Summarise(entropies)
}

func symbolLengthSummary(symbols []string) stats.SampleStatistics {
	lengths := utils.Transform(symbols, func(s string) int { return len(s) })
	return stats.Summarise(lengths)

}

func compareSummary(t *testing.T, name string, expected, actual stats.SampleStatistics) {
	if !expected.Equals(actual, 1e-4) {
		t.Errorf("%s summary did not match.\nExpected: %v\nActual: %v\n", name, expected, actual)
	}
}

func testSignals(t *testing.T, signals Signals, stringLiterals, identifiers []string) {
	expectedStringEntropySummary := symbolEntropySummary(stringLiterals)
	expectedStringLengthSummary := symbolLengthSummary(stringLiterals)
	expectedIdentifierEntropySummary := symbolEntropySummary(identifiers)
	expectedIdentifierLengthSummary := symbolLengthSummary(identifiers)

	compareSummary(t, "String literal entropy", expectedStringEntropySummary, signals.StringEntropySummary)
	compareSummary(t, "String literal lengths", expectedStringLengthSummary, signals.StringLengthSummary)
	compareSummary(t, "Identifier entropy", expectedIdentifierEntropySummary, signals.IdentifierEntropySummary)
	compareSummary(t, "Identifier lengths", expectedIdentifierLengthSummary, signals.IdentifierLengthSummary)

	expectedStringCombinedEntropy := stringentropy.CalculateEntropy(strings.Join(stringLiterals, ""), nil)
	if !utils.FloatEquals(expectedStringCombinedEntropy, signals.CombinedStringEntropy, 1e-4) {
		t.Errorf("Combined string entropy: expected %.3f, actual %.3f",
			expectedStringCombinedEntropy, signals.CombinedStringEntropy)
	}

	expectedIdentifierCombinedEntropy := stringentropy.CalculateEntropy(strings.Join(identifiers, ""), nil)
	if !utils.FloatEquals(expectedIdentifierCombinedEntropy, signals.CombinedIdentifierEntropy, 1e-4) {
		t.Errorf("Combined identifier entropy: expected %.3f, actual %.3f",
			expectedIdentifierCombinedEntropy, signals.CombinedIdentifierEntropy)
	}
}

type testCase struct {
	name           string
	jsSource       string
	stringLiterals []string
	identifiers    []string
}

var test1 = testCase{
	name: "simple 1",
	jsSource: `
var a = "hello"
	`,
	stringLiterals: []string{"hello"},
	identifiers:    []string{"a"},
}

var test2 = testCase{
	name: "simple 2",
	jsSource: `
function test(a, b = 2) {
	console.log("hello")
	var c = a + b
	if (c === 3) {
		return 4
	} else {
		return "apple"
	}
}
	`,
	stringLiterals: []string{"hello", "apple"},
	identifiers:    []string{"test", "a", "b", "c"},
}

func TestComputeSignals(t *testing.T) {
	const jsParserInstallDir = "/tmp/package-analysis-compute-signals-test-js-parser"
	parserConfig, err := js.InitParser(jsParserInstallDir)
	if err != nil {
		t.Errorf("failed to init parser: %v", err)
		return
	}

	testCases := []testCase{test1, test2}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			rawData, err := CollectData(parserConfig, "", test.jsSource, true)
			if err != nil {
				t.Error(err)
			} else {
				signals := ComputeSignals(*rawData)
				testSignals(t, signals, test.stringLiterals, test.identifiers)
			}
		})
	}

	if err = os.RemoveAll(jsParserInstallDir); err != nil {
		t.Errorf("failed to remove parser install dir: %v", err)
	}
}
