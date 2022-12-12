package obfuscation

import (
	"reflect"
	"testing"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing/js"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

type testCase struct {
	name         string
	jsSource     string
	expectedData RawData
}

var test1 = testCase{
	name: "simple 1",
	jsSource: `
var a = "hello"
	`,
	expectedData: RawData{
		Identifiers: []token.Identifier{
			{Name: "a", Type: parsing.Variable},
		},
		StringLiterals: []token.String{
			{Value: "hello", Raw: `"hello"`},
		},
		IntLiterals:   []token.Int{},
		FloatLiterals: []token.Float{},
	},
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
	expectedData: RawData{
		Identifiers: []token.Identifier{
			{Name: "test", Type: parsing.Function},
			{Name: "a", Type: parsing.Parameter},
			{Name: "b", Type: parsing.Parameter},
			{Name: "c", Type: parsing.Variable},
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
}

func init() {
	log.Initialize("")
}

func TestCollectData(t *testing.T) {
	parserConfig, err := js.InitParser(t.TempDir())
	if err != nil {
		t.Fatalf("failed to init parser: %v", err)
	}

	tests := []testCase{test1, test2}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CollectData(parserConfig, "", tt.jsSource, false)
			if err != nil {
				t.Errorf("%v", err)
				return
			}
			if !reflect.DeepEqual(got.Identifiers, tt.expectedData.Identifiers) {
				t.Errorf("Identifiers mismatch: got %v, want %v", got.Identifiers, tt.expectedData.Identifiers)
			}
			if !reflect.DeepEqual(got.StringLiterals, tt.expectedData.StringLiterals) {
				t.Errorf("String literals mismatch: got %v, want %v", got.StringLiterals, tt.expectedData.StringLiterals)
			}
			if !reflect.DeepEqual(got.IntLiterals, tt.expectedData.IntLiterals) {
				t.Errorf("Int literals mismatch: got %v, want %v", got.IntLiterals, tt.expectedData.IntLiterals)
			}
			if !reflect.DeepEqual(got.FloatLiterals, tt.expectedData.FloatLiterals) {
				t.Errorf("Float literals mismatch: got %v, want %v", got.FloatLiterals, tt.expectedData.FloatLiterals)
			}
		})
	}
}
