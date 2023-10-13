package staticanalysis

import (
	"reflect"
	"testing"

	"github.com/ossf/package-analysis/internal/staticanalysis/basicdata"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/signals"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis/token"
	"github.com/ossf/package-analysis/pkg/valuecounts"
)

func TestResult_ProduceSerializableResult(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		want   *staticanalysis.Results
	}{
		{
			name: "empty",
			result: Result{Files: []SingleResult{
				{
					Filename: "empty.txt",
					Basic:    &basicdata.FileData{},
					Parsing:  &parsing.SingleResult{},
					Signals:  &signals.FileSignals{},
				},
			}},
			want: &staticanalysis.Results{Files: []staticanalysis.FileResult{
				{
					Filename: "empty.txt",
				},
			}},
		},
		{
			name: "simple no js",
			result: Result{Files: []SingleResult{
				{
					Filename: "simple.txt",
					Basic: &basicdata.FileData{
						DetectedType: "plain text",
						Size:         10,
						SHA256:       "aabbbcc",
						LineLengths:  valuecounts.Count([]int{1, 2, 3, 4}),
					},
					Parsing: &parsing.SingleResult{},
					Signals: &signals.FileSignals{},
				},
			}},
			want: &staticanalysis.Results{Files: []staticanalysis.FileResult{
				{
					Filename:     "simple.txt",
					DetectedType: "plain text",
					Size:         10,
					SHA256:       "aabbbcc",
					LineLengths:  valuecounts.Count([]int{1, 2, 3, 4}),
				},
			}},
		},
		{
			name: "simple js",
			result: Result{Files: []SingleResult{
				{
					Filename: "simple.js",
					Basic: &basicdata.FileData{
						DetectedType: "javascript source file",
						Size:         100,
						SHA256:       "abc123def456",
						LineLengths:  valuecounts.Count([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
					},
					Parsing: &parsing.SingleResult{
						Language: parsing.JavaScript,
						Identifiers: []token.Identifier{
							{
								Name:    "myvar",
								Type:    token.Variable,
								Entropy: 0.5,
							},
							{
								Name:    "a",
								Type:    token.Variable,
								Entropy: 0.5,
							},
						},
						StringLiterals: []token.String{
							{
								Value:   "hello",
								Raw:     `"hello"`,
								Entropy: 0.4,
							},
							{
								Value:   "abcd",
								Raw:     `"\x61\x62\x63\x64"`,
								Entropy: 0.3,
							},
							{
								Value:   "https://github.com/ossf/package-analysis",
								Raw:     `"https://github.com/ossf/package-analysis"`,
								Entropy: 0.2,
							},
							{
								Value:   "192.168.0.1",
								Raw:     `192.168.0.1"`,
								Entropy: 0.25,
							},
						},
						IntLiterals: []token.Int{
							{
								Value: 10,
								Raw:   "10",
							},
						},
						FloatLiterals: []token.Float{
							{
								Value: 1.5,
								Raw:   "1.5",
							},
						},
						Comments: []token.Comment{
							{
								Text: "This is a comment",
							},
						},
					},
					Signals: &signals.FileSignals{
						IdentifierLengths: valuecounts.Count([]int{5}),
						StringLengths:     valuecounts.Count([]int{5}),
						SuspiciousIdentifiers: []staticanalysis.SuspiciousIdentifier{
							{
								Name: "a",
								Rule: "single",
							},
						},
						EscapedStrings: []staticanalysis.EscapedString{
							{
								Value:           "abcd",
								Raw:             `"\x61\x62\x63\x64"`,
								LevenshteinDist: 20,
							},
						},
						Base64Strings: []string{"abcd"},
						HexStrings:    []string{"abcd"},
						IPAddresses:   []string{"192.168.0.1"},
						URLs:          []string{"https://github.com/ossf/package-analysis"},
					},
				},
			}},
			want: &staticanalysis.Results{Files: []staticanalysis.FileResult{
				{
					Filename:     "simple.js",
					DetectedType: "javascript source file",
					Size:         100,
					SHA256:       "abc123def456",
					LineLengths:  valuecounts.Count([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
					Js: staticanalysis.JsData{
						Identifiers: []token.Identifier{
							{
								Name:    "myvar",
								Type:    token.Variable,
								Entropy: 0.5,
							},
							{
								Name:    "a",
								Type:    token.Variable,
								Entropy: 0.5,
							},
						},
						StringLiterals: []token.String{
							{
								Value:   "hello",
								Raw:     `"hello"`,
								Entropy: 0.4,
							},
							{
								Value:   "abcd",
								Raw:     `"\x61\x62\x63\x64"`,
								Entropy: 0.3,
							},
							{
								Value:   "https://github.com/ossf/package-analysis",
								Raw:     `"https://github.com/ossf/package-analysis"`,
								Entropy: 0.2,
							},
							{
								Value:   "192.168.0.1",
								Raw:     `192.168.0.1"`,
								Entropy: 0.25,
							},
						},
						IntLiterals: []token.Int{
							{
								Value: 10,
								Raw:   "10",
							},
						},
						FloatLiterals: []token.Float{
							{
								Value: 1.5,
								Raw:   "1.5",
							},
						},
						Comments: []token.Comment{
							{
								Text: "This is a comment",
							},
						},
					},
					IdentifierLengths: valuecounts.Count([]int{5}),
					StringLengths:     valuecounts.Count([]int{5}),
					SuspiciousIdentifiers: []staticanalysis.SuspiciousIdentifier{
						{
							Name: "a",
							Rule: "single",
						},
					},
					EscapedStrings: []staticanalysis.EscapedString{
						{
							Value:           "abcd",
							Raw:             `"\x61\x62\x63\x64"`,
							LevenshteinDist: 20,
						},
					},
					Base64Strings: []string{"abcd"},
					HexStrings:    []string{"abcd"},
					IPAddresses:   []string{"192.168.0.1"},
					URLs:          []string{"https://github.com/ossf/package-analysis"},
				},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.result
			got := r.ProduceSerializableResult()
			if !reflect.DeepEqual(got.Files[0].Base64Strings, tt.want.Files[0].Base64Strings) {
				t.Errorf("ProduceSerializableResult() base64 strings mismatch\n"+
					"got\n%v\nwant\n%v", got.Files[0].Base64Strings, tt.want.Files[0].Base64Strings)
			}
			if !reflect.DeepEqual(got.Files[0].HexStrings, tt.want.Files[0].HexStrings) {
				t.Errorf("ProduceSerializableResult() hex strings mismatch\n"+
					"got\n%v\nwant\n%v", got.Files[0].HexStrings, tt.want.Files[0].HexStrings)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProduceSerializableResult() mismatch\n"+
					"got\n%v\nwant\n%v", got, tt.want)
			}
		})
	}
}
