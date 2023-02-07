package obfuscation

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stats"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stringentropy"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
	"github.com/ossf/package-analysis/internal/utils"
)

type fileSignalsTestCase struct {
	name     string
	fileData parsing.SingleResult
}

var fileSignalsTestCases = []fileSignalsTestCase{
	{
		name: "simple 1",
		fileData: parsing.SingleResult{
			Identifiers: []token.Identifier{
				{Name: "a", Type: token.Variable},
			},
			StringLiterals: []token.String{
				{Value: "hello", Raw: `"hello"`},
			},
			IntLiterals:   []token.Int{},
			FloatLiterals: []token.Float{},
		},
	},
	{
		name: "simple 2",
		fileData: parsing.SingleResult{
			Identifiers: []token.Identifier{
				{Name: "test", Type: token.Function},
				{Name: "a", Type: token.Parameter},
				{Name: "b", Type: token.Parameter},
				{Name: "c", Type: token.Variable},
			},
			StringLiterals: []token.String{
				{Value: "hello", Raw: `"hello"`},
				{Value: "apple", Raw: `"apple"`},
			},
			IntLiterals: []token.Int{
				{Value: 2, Raw: "2"},
				{Value: 3, Raw: "3"},
				{Value: 4, Raw: "4"},
			},
			FloatLiterals: []token.Float{},
		},
	},
}

func symbolEntropySummary(symbols []string) stats.SampleStatistics {
	probs := stringentropy.CharacterProbabilities(symbols)
	entropies := utils.Transform(symbols, func(s string) float64 { return stringentropy.CalculateEntropy(s, probs) })
	return stats.Summarise(entropies)
}

func symbolLengthCounts(symbols []string) map[int]int {
	lengths := utils.Transform(symbols, func(s string) int { return len(s) })
	return stats.CountDistinct(lengths)
}

func compareSummary(t *testing.T, name string, expected, actual stats.SampleStatistics) {
	if !expected.Equals(actual, 1e-4) {
		t.Errorf("%s summary did not match.\nExpected: %v\nActual: %v\n", name, expected, actual)
	}
}

func compareCounts(t *testing.T, name string, expected, actual map[int]int) {
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("%s summary did not match.\nExpected: %v\nActual: %v\n", name, expected, actual)
	}
}

func testSignals(t *testing.T, signals FileSignals, stringLiterals []token.String, identifiers []token.Identifier) {
	literals := utils.Transform(stringLiterals, func(s token.String) string { return s.Value })
	identifierNames := utils.Transform(identifiers, func(i token.Identifier) string { return i.Name })
	expectedStringEntropySummary := symbolEntropySummary(literals)
	expectedStringLengthSummary := symbolLengthCounts(literals)
	expectedIdentifierEntropySummary := symbolEntropySummary(identifierNames)
	expectedIdentifierLengthSummary := symbolLengthCounts(identifierNames)

	compareSummary(t, "String literal entropy", expectedStringEntropySummary, signals.StringEntropySummary)
	compareCounts(t, "String literal lengths", expectedStringLengthSummary, signals.StringLengths)
	compareSummary(t, "Identifier entropy", expectedIdentifierEntropySummary, signals.IdentifierEntropySummary)
	compareCounts(t, "Identifier lengths", expectedIdentifierLengthSummary, signals.IdentifierLengths)

	expectedStringCombinedEntropy := stringentropy.CalculateEntropy(strings.Join(literals, ""), nil)
	if !utils.FloatEquals(expectedStringCombinedEntropy, signals.CombinedStringEntropy, 1e-4) {
		t.Errorf("Combined string entropy: expected %.3f, actual %.3f",
			expectedStringCombinedEntropy, signals.CombinedStringEntropy)
	}

	expectedIdentifierCombinedEntropy := stringentropy.CalculateEntropy(strings.Join(identifierNames, ""), nil)
	if !utils.FloatEquals(expectedIdentifierCombinedEntropy, signals.CombinedIdentifierEntropy, 1e-4) {
		t.Errorf("Combined identifier entropy: expected %.3f, actual %.3f",
			expectedIdentifierCombinedEntropy, signals.CombinedIdentifierEntropy)
	}
}

func init() {
	log.Initialize("")
}

func TestComputeSignals(t *testing.T) {
	for _, test := range fileSignalsTestCases {
		t.Run(test.name, func(t *testing.T) {
			signals := ComputeFileSignals(test.fileData)
			testSignals(t, signals, test.fileData.StringLiterals, test.fileData.Identifiers)
		})
	}
}
