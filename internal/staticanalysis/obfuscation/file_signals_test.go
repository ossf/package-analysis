package obfuscation

import (
	"reflect"
	"testing"

	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
	"github.com/ossf/package-analysis/internal/utils/valuecounts"
)

type fileSignalsTestCase struct {
	name            string
	parseData       parsing.SingleResult
	expectedSignals FileSignals
}

var fileSignalsTestCases = []fileSignalsTestCase{
	{
		name: "simple 1",
		parseData: parsing.SingleResult{
			Identifiers: []token.Identifier{
				{Name: "a", Type: token.Variable},
			},
			StringLiterals: []token.String{
				{Value: "hello", Raw: `"hello"`},
			},
			IntLiterals:   []token.Int{},
			FloatLiterals: []token.Float{},
		},
		expectedSignals: FileSignals{
			StringLengths:         valuecounts.ValueCounts{5: 1},
			IdentifierLengths:     valuecounts.ValueCounts{1: 1},
			SuspiciousIdentifiers: []SuspiciousIdentifier{{Name: "a", Rule: "single"}},
			EscapedStrings:        []EscapedString{},
			Base64Strings:         []string{},
			EmailAddresses:        []string{},
			HexStrings:            []string{},
			IPAddresses:           []string{},
			URLs:                  []string{},
		},
	},
	{
		name: "simple 2",
		parseData: parsing.SingleResult{
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
		expectedSignals: FileSignals{
			StringLengths:     valuecounts.ValueCounts{5: 2},
			IdentifierLengths: valuecounts.ValueCounts{1: 3, 4: 1},
			SuspiciousIdentifiers: []SuspiciousIdentifier{
				{Name: "a", Rule: "single"},
				{Name: "b", Rule: "single"},
				{Name: "c", Rule: "single"},
			},
			EscapedStrings: []EscapedString{},
			Base64Strings:  []string{},
			EmailAddresses: []string{},
			HexStrings:     []string{},
			IPAddresses:    []string{},
			URLs:           []string{},
		},
	},
}

func TestComputeSignals(t *testing.T) {
	for _, test := range fileSignalsTestCases {
		t.Run(test.name, func(t *testing.T) {
			signals := ComputeFileSignals(test.parseData)
			if !reflect.DeepEqual(signals, test.expectedSignals) {
				t.Errorf("actual signals did not match expected\n"+
					"== want ==\n%v\n== got ==\n%v\n======", test.expectedSignals, signals)
			}
		})
	}
}
