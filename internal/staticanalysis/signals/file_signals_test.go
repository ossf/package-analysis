package signals

import (
	"reflect"
	"testing"

	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis/token"
	"github.com/ossf/package-analysis/pkg/valuecounts"
)

type fileSignalsTestCase struct {
	name            string
	parseData       parsing.SingleResult
	expectedSignals FileSignals
}

var fileSignalsTestCases = []fileSignalsTestCase{
	{
		name:      "empty",
		parseData: parsing.SingleResult{},
		expectedSignals: FileSignals{
			StringLengths:         valuecounts.New(),
			IdentifierLengths:     valuecounts.New(),
			SuspiciousIdentifiers: []staticanalysis.SuspiciousIdentifier{},
			EscapedStrings:        []staticanalysis.EscapedString{},
			Base64Strings:         []string{},
			HexStrings:            []string{},
			IPAddresses:           []string{},
			URLs:                  []string{},
		},
	},
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
			StringLengths:         valuecounts.Count([]int{5}),
			IdentifierLengths:     valuecounts.Count([]int{1}),
			SuspiciousIdentifiers: []staticanalysis.SuspiciousIdentifier{{Name: "a", Rule: "single"}},
			EscapedStrings:        []staticanalysis.EscapedString{},
			Base64Strings:         []string{},
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
			StringLengths:     valuecounts.Count([]int{5, 5}),
			IdentifierLengths: valuecounts.Count([]int{4, 1, 1, 1}),
			SuspiciousIdentifiers: []staticanalysis.SuspiciousIdentifier{
				{Name: "a", Rule: "single"},
				{Name: "b", Rule: "single"},
				{Name: "c", Rule: "single"},
			},
			EscapedStrings: []staticanalysis.EscapedString{},
			Base64Strings:  []string{},
			HexStrings:     []string{},
			IPAddresses:    []string{},
			URLs:           []string{},
		},
	},
	{
		name: "one of everything",
		parseData: parsing.SingleResult{
			Identifiers: []token.Identifier{
				{Name: "_0x12414124", Type: token.Variable},
				{Name: "a", Type: token.Parameter},
				{Name: "d1912931", Type: token.Parameter},
			},
			StringLiterals: []token.String{
				{Value: "hello@email.me", Raw: `"hello@email.me"`},
				{Value: "https://this.is.a.website.com", Raw: `"https://this.is.a.website.com"`},
				{Value: "aGVsbG8gd29ybGQK", Raw: `"aGVsbG8gd29ybGQK"`},
				{Value: "8.8.8.8", Raw: `"8.8.8.8"`},
				{Value: "e3fc:234a:2341::abcd", Raw: `"e3fc:234a:2341::abcd"`},
				{Value: "0x21323492394", Raw: `"0x21323492394"`},
			},
			IntLiterals:   []token.Int{},
			FloatLiterals: []token.Float{},
		},
		expectedSignals: FileSignals{
			IdentifierLengths: valuecounts.Count([]int{11, 1, 8}),
			StringLengths:     valuecounts.Count([]int{14, 29, 16, 7, 20, 13}),
			SuspiciousIdentifiers: []staticanalysis.SuspiciousIdentifier{
				{Name: "_0x12414124", Rule: "hex"},
				{Name: "a", Rule: "single"},
				{Name: "d1912931", Rule: "numeric"},
			},
			EscapedStrings: []staticanalysis.EscapedString{},
			Base64Strings:  []string{"aGVsbG8gd29ybGQK"},
			HexStrings:     []string{"21323492394"},
			IPAddresses:    []string{"8.8.8.8", "e3fc:234a:2341::abcd"},
			URLs:           []string{"https://this.is.a.website.com"},
		},
	},
	{
		name: "escaped strings",
		parseData: parsing.SingleResult{
			StringLiterals: []token.String{
				{Value: "@ABCD", Raw: "\\100\\101\\102\\103\\104"},
				{Value: "@ABCD", Raw: "\\x40\\x41\\x42\\x43\\x44"},
				{Value: "@ABCD", Raw: "\\u0040\\u0041\\u0042\\u0043\\u0044"},
				{Value: "@ABCD", Raw: "\\U00000040\\U00000041\\U00000042\\U00000043\\U00000044"},
			},
		},
		expectedSignals: FileSignals{
			StringLengths:         valuecounts.Count([]int{5, 5, 5, 5}),
			IdentifierLengths:     valuecounts.New(),
			SuspiciousIdentifiers: []staticanalysis.SuspiciousIdentifier{},
			Base64Strings:         []string{},
			HexStrings:            []string{},
			IPAddresses:           []string{},
			URLs:                  []string{},
			EscapedStrings: []staticanalysis.EscapedString{
				{Value: "@ABCD", Raw: "\\100\\101\\102\\103\\104", LevenshteinDist: 25},
				{Value: "@ABCD", Raw: "\\x40\\x41\\x42\\x43\\x44", LevenshteinDist: 25},
				{Value: "@ABCD", Raw: "\\u0040\\u0041\\u0042\\u0043\\u0044", LevenshteinDist: 35},
				{Value: "@ABCD", Raw: "\\U00000040\\U00000041\\U00000042\\U00000043\\U00000044", LevenshteinDist: 55},
			},
		},
	},
}

func TestComputeSignals(t *testing.T) {
	for _, test := range fileSignalsTestCases {
		t.Run(test.name, func(t *testing.T) {
			signals := AnalyzeSingle(test.parseData)
			if !reflect.DeepEqual(signals, test.expectedSignals) {
				t.Errorf("actual signals did not match expected\n"+
					"== want ==\n%v\n== got ==\n%#v\n======", test.expectedSignals, signals)
			}
		})
	}
}
