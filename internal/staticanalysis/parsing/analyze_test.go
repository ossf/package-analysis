package parsing

import (
	"context"
	"math"
	"reflect"
	"testing"

	"github.com/ossf/package-analysis/internal/staticanalysis/externalcmd"
	"github.com/ossf/package-analysis/internal/staticanalysis/signals/stringentropy"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis/token"
)

type analyzeTestcase struct {
	name         string
	jsSource     string
	expectedData SingleResult
}

var literalCharProbs = []map[rune]float64{
	stringentropy.CharacterProbabilities([]string{"hello"}),
	stringentropy.CharacterProbabilities([]string{"hello", "apple"}),
}
var identifierCharProbs = []map[rune]float64{
	stringentropy.CharacterProbabilities([]string{"a"}),
	stringentropy.CharacterProbabilities([]string{"test", "a", "b", "c"}),
}

var analyzeTestcases = []analyzeTestcase{
	{
		name:     "console log hi",
		jsSource: `console.log("hi");`,
		expectedData: SingleResult{
			Language:    JavaScript,
			Identifiers: []token.Identifier{
				// Members excluded
				//{Name: "log", Type: token.Member, Entropy: math.Log(3)},
			},
			StringLiterals: []token.String{
				{Value: "hi", Raw: `"hi"`, Entropy: math.Log(2)},
			},
			IntLiterals:   []token.Int{},
			FloatLiterals: []token.Float{},
			Comments:      []token.Comment{},
		},
	},
	{
		name: "simple 1",
		jsSource: `
var a = "hello"
	`,
		expectedData: SingleResult{
			Language: JavaScript,
			Identifiers: []token.Identifier{
				{Name: "a", Type: token.Variable, Entropy: stringentropy.Calculate("a", identifierCharProbs[0])},
			},
			StringLiterals: []token.String{
				{Value: "hello", Raw: `"hello"`, Entropy: stringentropy.Calculate("hello", literalCharProbs[0])},
			},
			IntLiterals:   []token.Int{},
			FloatLiterals: []token.Float{},
			Comments:      []token.Comment{},
		},
	},
	{
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
		expectedData: SingleResult{
			Language: JavaScript,
			Identifiers: []token.Identifier{
				{Name: "test", Type: token.Function, Entropy: stringentropy.Calculate("test", identifierCharProbs[1])},
				{Name: "a", Type: token.Parameter, Entropy: stringentropy.Calculate("a", identifierCharProbs[1])},
				{Name: "b", Type: token.Parameter, Entropy: stringentropy.Calculate("b", identifierCharProbs[1])},
				{Name: "c", Type: token.Variable, Entropy: stringentropy.Calculate("c", identifierCharProbs[1])},
			},
			StringLiterals: []token.String{
				{Value: "hello", Raw: `"hello"`, Entropy: stringentropy.Calculate("hello", literalCharProbs[1])},
				{Value: "apple", Raw: `"apple"`, Entropy: stringentropy.Calculate("apple", literalCharProbs[1])},
			},
			IntLiterals: []token.Int{
				{Value: 2, Raw: "2"},
				{Value: 3, Raw: "3"},
				{Value: 4, Raw: "4"},
			},
			FloatLiterals: []token.Float{},
			Comments:      []token.Comment{},
		},
	},
	{
		name:     "invalid 1",
		jsSource: "this is not JavaScript",
		expectedData: SingleResult{
			Language:       NoLanguage,
			Identifiers:    []token.Identifier{},
			StringLiterals: []token.String{},
			IntLiterals:    []token.Int{},
			FloatLiterals:  []token.Float{},
			Comments:       []token.Comment{},
		},
	},
}

func TestAnalyze(t *testing.T) {
	parserConfig, err := InitParser(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("failed to init parser: %v", err)
	}

	for _, tt := range analyzeTestcases {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Analyze(context.Background(), parserConfig, externalcmd.StringInput(tt.jsSource), false)
			if err != nil {
				t.Errorf("%v", err)
				return
			}
			got := result["stdin"]

			if got.Language != tt.expectedData.Language {
				t.Errorf("Filename mismatch: got %s, want %s", got.Language, tt.expectedData.Language)
			}

			if !reflect.DeepEqual(got.Identifiers, tt.expectedData.Identifiers) {
				t.Errorf("Identifiers mismatch: got %#v, want %v", got.Identifiers, tt.expectedData.Identifiers)
			}
			if !reflect.DeepEqual(got.StringLiterals, tt.expectedData.StringLiterals) {
				t.Errorf("String literals mismatch: got %#v, want %v", got.StringLiterals, tt.expectedData.StringLiterals)
			}
			if !reflect.DeepEqual(got.IntLiterals, tt.expectedData.IntLiterals) {
				t.Errorf("Int literals mismatch: got %#v, want %v", got.IntLiterals, tt.expectedData.IntLiterals)
			}
			if !reflect.DeepEqual(got.FloatLiterals, tt.expectedData.FloatLiterals) {
				t.Errorf("Float literals mismatch: got %#v, want %v", got.FloatLiterals, tt.expectedData.FloatLiterals)
			}
			if !reflect.DeepEqual(got.Comments, tt.expectedData.Comments) {
				t.Errorf("Comments mismatch: got %#v, want %v", got.Comments, tt.expectedData.Comments)
			}
		})
	}
}
