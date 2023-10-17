package parsing

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"testing"

	"github.com/ossf/package-analysis/internal/staticanalysis/externalcmd"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis/token"
)

type jsTestCase struct {
	name      string
	inputJS   string
	want      singleParseData
	printJSON bool // set to true to see raw parser output
}

var jsTestCases = []jsTestCase{
	{
		name: "test string declarations and templates",
		inputJS: `
function test() {
    var mystring1 = "hello1";
    var mystring2 = 'hello2';
    var mystring3 = "hello'3'";
    var mystring4 = 'hello"4"';
    var mystring5 = "hello\"5\"";
    var mystring6 = "hello\'6\'";
    var mystring7 = 'hello\'7\'';
    var mystring8 = "hello" + "8";
    var mystring9 = ` + "`hello9`" + `;
    var mystring10 = ` + "`hello\"'${10}\"'`" + `;
	var mystring11 = ` + "`hello" + `
//"'11"'` + "`" + `;
	var mystring12 = ` + "`hello\"'${5.6 + 6.4}\"'`" + `;
}`,
		want: singleParseData{
			Identifiers: []parsedIdentifier{
				{token.Function, "test", token.Position{2, 9}},
				{token.Variable, "mystring1", token.Position{3, 8}},
				{token.Variable, "mystring2", token.Position{4, 8}},
				{token.Variable, "mystring3", token.Position{5, 8}},
				{token.Variable, "mystring4", token.Position{6, 8}},
				{token.Variable, "mystring5", token.Position{7, 8}},
				{token.Variable, "mystring6", token.Position{8, 8}},
				{token.Variable, "mystring7", token.Position{9, 8}},
				{token.Variable, "mystring8", token.Position{10, 8}},
				{token.Variable, "mystring9", token.Position{11, 8}},
				{token.Variable, "mystring10", token.Position{12, 8}},
				{token.Variable, "mystring11", token.Position{13, 5}},
				{token.Variable, "mystring12", token.Position{15, 5}},
			},
			Literals: []parsedLiteral[any]{
				{"String", "string", "hello1", `"hello1"`, false, token.Position{3, 20}},
				{"String", "string", "hello2", `'hello2'`, false, token.Position{4, 20}},
				{"String", "string", "hello'3'", `"hello'3'"`, false, token.Position{5, 20}},
				{"String", "string", "hello\"4\"", `'hello"4"'`, false, token.Position{6, 20}},
				{"String", "string", "hello\"5\"", `"hello\"5\""`, false, token.Position{7, 20}},
				{"String", "string", "hello'6'", `"hello\'6\'"`, false, token.Position{8, 20}},
				{"String", "string", "hello'7'", `'hello\'7\''`, false, token.Position{9, 20}},
				{"String", "string", "hello", `"hello"`, false, token.Position{10, 20}},
				{"String", "string", "8", `"8"`, false, token.Position{10, 30}},
				{"StringTemplate", "string", "hello9", "`hello9`", false, token.Position{11, 20}},
				{"StringTemplate", "string", "hello\"'${}\"'", "`hello\"'${}\"'`", false, token.Position{12, 21}},
				{"Numeric", "float64", 10.0, "10", false, token.Position{12, 31}},
				{"StringTemplate", "string", "hello\n//\"'11\"'", "`hello\n//\"'11\"'`", false, token.Position{13, 18}},
				{"StringTemplate", "string", "hello\"'${}\"'", "`hello\"'${}\"'`", false, token.Position{15, 18}},
				{"Numeric", "float64", 5.6, "5.6", false, token.Position{15, 28}},
				{"Numeric", "float64", 6.4, "6.4", false, token.Position{15, 34}},
			},
		},
	},
	{
		name: "test function parameters",
		inputJS: `
function test2(param1, param2, param3 = "ahd") {
	return param1 + param2 + param3;
}`,
		want: singleParseData{
			Identifiers: []parsedIdentifier{
				{token.Function, "test2", token.Position{2, 9}},
				{token.Parameter, "param1", token.Position{2, 15}},
				{token.Parameter, "param2", token.Position{2, 23}},
				{token.Parameter, "param3", token.Position{2, 31}},
			},
			Literals: []parsedLiteral[any]{
				{"String", "string", "ahd", `"ahd"`, false, token.Position{2, 40}},
			},
		},
	},
	{
		name: "test control flow",
		inputJS: `
function test3(a, b, c) {
    for (var i = a; i < b; i++) {
outer:
        for (var j = 1; j < 3; j++) {
            for (var k = j; k < j + 10; k++) {
                if (j === 2) {
                    break outer;
                }
            }
        }
        c = c * i;
        if (c % 32 === 0) {
            continue;
        }
        console.log("here");
    }
    console.log("End");
}`,
		want: singleParseData{
			Identifiers: []parsedIdentifier{
				{token.Function, "test3", token.Position{2, 9}},
				{token.Parameter, "a", token.Position{2, 15}},
				{token.Parameter, "b", token.Position{2, 18}},
				{token.Parameter, "c", token.Position{2, 21}},
				{token.Variable, "i", token.Position{3, 13}},
				{token.StatementLabel, "outer", token.Position{4, 0}},
				{token.Variable, "j", token.Position{5, 17}},
				{token.Variable, "k", token.Position{6, 21}},
				{token.Member, "log", token.Position{16, 16}},
				{token.Member, "log", token.Position{18, 12}},
			},
			Literals: []parsedLiteral[any]{
				{"Numeric", "float64", 1.0, "1", false, token.Position{5, 21}},
				{"Numeric", "float64", 3.0, "3", false, token.Position{5, 28}},
				{"Numeric", "float64", 10.0, "10", false, token.Position{6, 36}},
				{"Numeric", "float64", 2.0, "2", false, token.Position{7, 26}},
				{"Numeric", "float64", 32.0, "32", false, token.Position{13, 16}},
				{"Numeric", "float64", 0.0, "0", false, token.Position{13, 23}},
				{"String", "string", "here", `"here"`, false, token.Position{16, 20}},
				{"String", "string", "End", `"End"`, false, token.Position{18, 16}},
			},
		},
	},
	{
		name: "test arrays and exceptions",
		inputJS: `
function test4() {
    const a = [1, 2, 3];
    try {
        if (a[1] === 3) {
            console.log(a[-1]); // NB the literal here is 1, not -1!
        } else if (a[1] === 2) {
            console.log(a[1]);
        } else {
            console.log(a[2]);
        }
    } catch (e) {
        var f = "abc";
        console.log(e + f);
    }

    switch (a[0]) {
        case 1:
            console.log("Hp");
            break;
        default:
            console.log("Hq");
            break;
    }
}`,
		want: singleParseData{
			Identifiers: []parsedIdentifier{
				{token.Function, "test4", token.Position{2, 9}},
				{token.Variable, "a", token.Position{3, 10}},
				{token.Member, "log", token.Position{6, 20}},
				{token.Member, "log", token.Position{8, 20}},
				{token.Member, "log", token.Position{10, 20}},
				{token.Parameter, "e", token.Position{12, 13}},
				{token.Variable, "f", token.Position{13, 12}},
				{token.Member, "log", token.Position{14, 16}},
				{token.Member, "log", token.Position{19, 20}},
				{token.Member, "log", token.Position{22, 20}},
			},
			Literals: []parsedLiteral[any]{
				{"Numeric", "float64", 1.0, "1", true, token.Position{3, 15}},
				{"Numeric", "float64", 2.0, "2", true, token.Position{3, 18}},
				{"Numeric", "float64", 3.0, "3", true, token.Position{3, 21}},
				{"Numeric", "float64", 1.0, "1", false, token.Position{5, 14}},
				{"Numeric", "float64", 3.0, "3", false, token.Position{5, 21}},
				{"Numeric", "float64", 1.0, "1", false, token.Position{6, 27}},
				{"Numeric", "float64", 1.0, "1", false, token.Position{7, 21}},
				{"Numeric", "float64", 2.0, "2", false, token.Position{7, 28}},
				{"Numeric", "float64", 1.0, "1", false, token.Position{8, 26}},
				{"Numeric", "float64", 2.0, "2", false, token.Position{10, 26}},
				{"String", "string", "abc", `"abc"`, false, token.Position{13, 16}},
				{"Numeric", "float64", 0.0, "0", false, token.Position{17, 14}},
				{"Numeric", "float64", 1.0, "1", false, token.Position{18, 13}},
				{"String", "string", "Hp", `"Hp"`, false, token.Position{19, 24}},
				{"String", "string", "Hq", `"Hq"`, false, token.Position{22, 24}},
			},
		},
	},
	{
		name: "test class definition",
		inputJS: `
// unnamed
let Rectangle = class {
    constructor(height, width) {
        this.height = height;
        this.width = width;
    }
};
console.log(Rectangle.name);
// output: "Rectangle"

// named
Rectangle = class Rectangle2 {
    #test = false;
    constructor(height, width) {
        this.height = height;
        this.width = width;
    }
};
console.log(Rectangle.name);
// output: "Rectangle2"
`,
		want: singleParseData{
			Identifiers: []parsedIdentifier{
				{token.Variable, "Rectangle", token.Position{3, 4}},
				{token.Parameter, "height", token.Position{4, 16}},
				{token.Parameter, "width", token.Position{4, 24}},
				{token.Member, "height", token.Position{5, 13}},
				{token.Member, "width", token.Position{6, 13}},
				{token.Member, "log", token.Position{9, 8}},
				{token.Member, "name", token.Position{9, 22}},
				// {Variable, "Rectangle", Position{4, 22}},
				{token.Class, "Rectangle2", token.Position{13, 18}},
				{token.Property, "test", token.Position{14, 5}},
				{token.Parameter, "height", token.Position{15, 16}},
				{token.Parameter, "width", token.Position{15, 24}},
				{token.Member, "height", token.Position{16, 13}},
				{token.Member, "width", token.Position{17, 13}},
				{token.Member, "log", token.Position{20, 8}},
				{token.Member, "name", token.Position{20, 22}},
			},
			Literals: []parsedLiteral[any]{},
		},
	},
	{
		name: "test use strict",
		inputJS: `
'use strict';
console.log("Hello");
`,
		want: singleParseData{
			Identifiers: []parsedIdentifier{
				{token.Member, "log", token.Position{3, 8}},
			},
			Literals: []parsedLiteral[any]{
				{"String", "string", "use strict", `'use strict'`, false, token.Position{2, 0}},
				{"String", "string", "Hello", `"Hello"`, false, token.Position{3, 12}},
			},
		},
	},
	{
		name: "test exotic assignments",
		inputJS: `
let [a, b] = [1, 2];
let [_, c] = [3, 4];
var index = 0,
    completed = 0,
    {length, width} = 10,
    cancelled = false;
`,
		want: singleParseData{
			Identifiers: []parsedIdentifier{
				{token.Variable, "a", token.Position{2, 5}},
				{token.Variable, "b", token.Position{2, 8}},
				{token.Variable, "_", token.Position{3, 5}},
				{token.Variable, "c", token.Position{3, 8}},
				{token.Variable, "index", token.Position{4, 4}},
				{token.Variable, "completed", token.Position{5, 4}},
				{token.Variable, "length", token.Position{6, 5}},
				{token.Variable, "width", token.Position{6, 13}},
				{token.Variable, "cancelled", token.Position{7, 4}},
			},
			Literals: []parsedLiteral[any]{
				{"Numeric", "float64", 1.0, "1", true, token.Position{2, 14}},
				{"Numeric", "float64", 2.0, "2", true, token.Position{2, 17}},
				{"Numeric", "float64", 3.0, "3", true, token.Position{3, 14}},
				{"Numeric", "float64", 4.0, "4", true, token.Position{3, 17}},
				{"Numeric", "float64", 0.0, "0", false, token.Position{4, 12}},
				{"Numeric", "float64", 0.0, "0", false, token.Position{5, 16}},
				{"Numeric", "float64", 10.0, "10", false, token.Position{6, 22}},
			},
		},
	},
	{
		name: "test regex literal",
		inputJS: `
function validateIPAddress(ipaddress) {
	const regex = /(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)/
    if (regex.test(ipaddress) || ipaddress.toLowerCase().includes('localhost')) {
        return (true)
    }

    return (false)
}
`,
		want: singleParseData{
			ValidInput: true,
			Identifiers: []parsedIdentifier{
				{token.Function, "validateIPAddress", token.Position{2, 9}},
				{token.Parameter, "ipaddress", token.Position{2, 27}},
				{token.Variable, "regex", token.Position{3, 7}},
				{token.Member, "test", token.Position{4, 14}},
				{token.Member, "toLowerCase", token.Position{4, 43}},
				{token.Member, "includes", token.Position{4, 57}},
			},
			Literals: []parsedLiteral[any]{
				{
					Type:     "Regexp",
					GoType:   "string",
					Value:    "(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)",
					RawValue: "/(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)/",
					InArray:  false,
					Pos:      token.Position{3, 15},
				},
				{"String", "string", "localhost", "'localhost'", false, token.Position{4, 66}},
			},
		},
		printJSON: true,
	},
	{
		name: "test big integers",
		inputJS: `
let a = 123456789123456789n;     // 123456789123456789
let b = 0o777777777777n;         // 68719476735
let c = 0x123456789ABCDEFn;      // 81985529216486895
let d = 0b11101001010101010101n; // 955733
`,
		want: singleParseData{
			ValidInput: true,
			Identifiers: []parsedIdentifier{
				{token.Variable, "a", token.Position{2, 4}},
				{token.Variable, "b", token.Position{3, 4}},
				{token.Variable, "c", token.Position{4, 4}},
				{token.Variable, "d", token.Position{5, 4}},
			},
			Literals: []parsedLiteral[any]{
				{"Numeric", "big.Int", big.NewInt(123456789123456789), "123456789123456789n", false, token.Position{2, 8}},
				{"Numeric", "big.Int", big.NewInt(68719476735), "0o777777777777n", false, token.Position{3, 8}},
				{"Numeric", "big.Int", big.NewInt(81985529216486895), "0x123456789ABCDEFn", false, token.Position{4, 8}},
				{"Numeric", "big.Int", big.NewInt(955733), "0b11101001010101010101n", false, token.Position{5, 8}},
			},
		},
		printJSON: false,
	},
	{
		name: "test syntax error",
		inputJS: `
a = w w;
`,
		want: singleParseData{
			ValidInput:  false,
			Identifiers: []parsedIdentifier{},
			Errors: []parserStatus{
				{
					"Error", "SyntaxError", "BABEL_PARSER_SYNTAX_ERROR: MissingSemicolon", token.Position{2, 5},
				},
				{
					"Error", "SyntaxError", "FATAL SYNTAX ERROR (unable to parse remainder of file)", token.Position{2, 5},
				},
			},
		},
		printJSON: false,
	},
	{
		name: "more string templates",
		inputJS: "console.log(`the operation ${1} \\u2297 ${2} equals ${5}`);\n" +
			"console.log(`\\u{54}\\u0065\\x78t`);",
		want: singleParseData{
			ValidInput: true,
			Identifiers: []parsedIdentifier{
				{token.Member, "log", token.Position{1, 8}},
				{token.Member, "log", token.Position{2, 8}},
			},
			Literals: []parsedLiteral[any]{
				{"StringTemplate", "string", "the operation ${} ⊗ ${} equals ${}",
					"`the operation ${} \\u2297 ${} equals ${}`", false, token.Position{1, 12}},
				{"Numeric", "float64", 1.0, "1", false, token.Position{1, 29}},
				{"Numeric", "float64", 2.0, "2", false, token.Position{1, 41}},
				{"Numeric", "float64", 5.0, "5", false, token.Position{1, 53}},
				{"StringTemplate", "string", "Text", "`\\u{54}\\u0065\\x78t`", false, token.Position{2, 12}},
			},
		},

		printJSON: false,
	},
}

func TestParseJS(t *testing.T) {
	const printAllJSON = false

	jsParserConfig, err := InitParser(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("%v", err)
	}

	for _, tt := range jsTestCases {
		t.Run(tt.name, func(t *testing.T) {
			result, rawOutput, err := parseJS(context.Background(), jsParserConfig, externalcmd.StringInput(tt.inputJS))
			got := result["stdin"]
			if err != nil {
				t.Errorf("parseJS() error = %v", err)
				fmt.Println("Parser output:\n", rawOutput)
				return
			}
			if len(tt.want.Literals) != len(got.Literals) {
				t.Errorf("Mismatch in number of literals: want %d, got %d", len(tt.want.Literals), len(got.Literals))
			}
			for i, wantLiteral := range tt.want.Literals {
				if i >= len(got.Literals) {
					t.Errorf("Literal missing: want %v", wantLiteral)
				} else {
					gotLiteral := got.Literals[i]
					if !reflect.DeepEqual(gotLiteral, wantLiteral) {
						t.Errorf("Literals mismatch (#%d):\ngot  %v\nwant %v", i+1, gotLiteral, wantLiteral)
					}
				}
			}

			if len(tt.want.Identifiers) != len(got.Identifiers) {
				t.Errorf("Mismatch in number of identifiers: want %d, got %d", len(tt.want.Identifiers), len(got.Identifiers))
			}
			for i, wantIdentifier := range tt.want.Identifiers {
				if i >= len(got.Identifiers) {
					t.Errorf("Identifier missing: want %v", wantIdentifier)
				} else {
					gotIdentifier := got.Identifiers[i]
					if !reflect.DeepEqual(gotIdentifier, wantIdentifier) {
						t.Errorf("Identifier mismatch (#%d):\ngot  %v\nwant %v", i+1, gotIdentifier, wantIdentifier)
					}
				}
			}

			if len(tt.want.Errors) != len(got.Errors) {
				t.Errorf("Mismatch in number of errors: want %d, got %d", len(tt.want.Errors), len(got.Errors))
			}
			if len(tt.want.Errors) > 0 {
				for i, wantErr := range tt.want.Errors {
					if i >= len(got.Errors) {
						t.Errorf("Error missing: want %v", wantErr)
					} else if got.Errors[i] != wantErr {
						t.Errorf("Error mismatch (#%d):\ngot  %v\nwant %v", i+1, got.Errors[i], wantErr)
					}
				}
			}

			if t.Failed() || printAllJSON {
				fmt.Println("Raw JSON:\n", rawOutput)
			}
		})
	}
}
