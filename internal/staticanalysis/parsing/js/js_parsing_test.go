package js

import (
	"reflect"
	"testing"

	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
)

const jsParserPath = "./babel-parser.js"

type jsTestCase struct {
	name      string
	inputJs   string
	want      parsing.ParseResult
	printJson bool // set to true to see raw parser output
}

var test1 = jsTestCase{
	name: "test string declarations and templates",
	inputJs: `
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
	want: parsing.ParseResult{
		Identifiers: []parsing.ParsedIdentifier{
			{parsing.Function, "test", parsing.TextPosition{2, 9}},
			{parsing.Variable, "mystring1", parsing.TextPosition{3, 8}},
			{parsing.Variable, "mystring2", parsing.TextPosition{4, 8}},
			{parsing.Variable, "mystring3", parsing.TextPosition{5, 8}},
			{parsing.Variable, "mystring4", parsing.TextPosition{6, 8}},
			{parsing.Variable, "mystring5", parsing.TextPosition{7, 8}},
			{parsing.Variable, "mystring6", parsing.TextPosition{8, 8}},
			{parsing.Variable, "mystring7", parsing.TextPosition{9, 8}},
			{parsing.Variable, "mystring8", parsing.TextPosition{10, 8}},
			{parsing.Variable, "mystring9", parsing.TextPosition{11, 8}},
			{parsing.Variable, "mystring10", parsing.TextPosition{12, 8}},
			{parsing.Variable, "mystring11", parsing.TextPosition{13, 5}},
			{parsing.Variable, "mystring12", parsing.TextPosition{15, 5}},
		},
		Literals: []parsing.ParsedLiteral[any]{
			{"String", "string", "hello1", `"hello1"`, false, parsing.TextPosition{3, 20}},
			{"String", "string", "hello2", `'hello2'`, false, parsing.TextPosition{4, 20}},
			{"String", "string", "hello'3'", `"hello'3'"`, false, parsing.TextPosition{5, 20}},
			{"String", "string", "hello\"4\"", `'hello"4"'`, false, parsing.TextPosition{6, 20}},
			{"String", "string", "hello\"5\"", `"hello\"5\""`, false, parsing.TextPosition{7, 20}},
			{"String", "string", "hello'6'", `"hello\'6\'"`, false, parsing.TextPosition{8, 20}},
			{"String", "string", "hello'7'", `'hello\'7\''`, false, parsing.TextPosition{9, 20}},
			{"String", "string", "hello", `"hello"`, false, parsing.TextPosition{10, 20}},
			{"String", "string", "8", `"8"`, false, parsing.TextPosition{10, 30}},
			{"StringTemplate", "string", "hello9", `hello9`, false, parsing.TextPosition{11, 21}},
			{"StringTemplate", "string", "hello\"'", `hello"'`, false, parsing.TextPosition{12, 22}},
			{"StringTemplate", "string", "\"'", `"'`, false, parsing.TextPosition{12, 34}},
			{"Numeric", "float64", 10.0, "10", false, parsing.TextPosition{12, 31}},
			{"StringTemplate", "string", "hello\n//\"'11\"'", `hello` + "\n" + `//"'11"'`, false, parsing.TextPosition{13, 19}},
			{"StringTemplate", "string", "hello\"'", `hello"'`, false, parsing.TextPosition{15, 19}},
			{"StringTemplate", "string", "\"'", `"'`, false, parsing.TextPosition{15, 38}},
			{"Numeric", "float64", 5.6, "5.6", false, parsing.TextPosition{15, 28}},
			{"Numeric", "float64", 6.4, "6.4", false, parsing.TextPosition{15, 34}},
		},
	},
}

var test2 = jsTestCase{
	name: "test function parameters",
	inputJs: `
function test2(param1, param2, param3 = "ahd") {
	return param1 + param2 + param3;
}`,
	want: parsing.ParseResult{
		Identifiers: []parsing.ParsedIdentifier{
			{parsing.Function, "test2", parsing.TextPosition{2, 9}},
			{parsing.Parameter, "param1", parsing.TextPosition{2, 15}},
			{parsing.Parameter, "param2", parsing.TextPosition{2, 23}},
			{parsing.Parameter, "param3", parsing.TextPosition{2, 31}},
		},
		Literals: []parsing.ParsedLiteral[any]{
			{"String", "string", "ahd", `"ahd"`, false, parsing.TextPosition{2, 40}},
		},
	},
}

var test3 = jsTestCase{
	name: "test control flow",
	inputJs: `
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
	want: parsing.ParseResult{
		Identifiers: []parsing.ParsedIdentifier{
			{parsing.Function, "test3", parsing.TextPosition{2, 9}},
			{parsing.Parameter, "a", parsing.TextPosition{2, 15}},
			{parsing.Parameter, "b", parsing.TextPosition{2, 18}},
			{parsing.Parameter, "c", parsing.TextPosition{2, 21}},
			{parsing.Variable, "i", parsing.TextPosition{3, 13}},
			{parsing.StatementLabel, "outer", parsing.TextPosition{4, 0}},
			{parsing.Variable, "j", parsing.TextPosition{5, 17}},
			{parsing.Variable, "k", parsing.TextPosition{6, 21}},
			{parsing.Member, "log", parsing.TextPosition{16, 16}},
			{parsing.Member, "log", parsing.TextPosition{18, 12}},
		},
		Literals: []parsing.ParsedLiteral[any]{
			{"Numeric", "float64", 1.0, "1", false, parsing.TextPosition{5, 21}},
			{"Numeric", "float64", 3.0, "3", false, parsing.TextPosition{5, 28}},
			{"Numeric", "float64", 10.0, "10", false, parsing.TextPosition{6, 36}},
			{"Numeric", "float64", 2.0, "2", false, parsing.TextPosition{7, 26}},
			{"Numeric", "float64", 32.0, "32", false, parsing.TextPosition{13, 16}},
			{"Numeric", "float64", 0.0, "0", false, parsing.TextPosition{13, 23}},
			{"String", "string", "here", `"here"`, false, parsing.TextPosition{16, 20}},
			{"String", "string", "End", `"End"`, false, parsing.TextPosition{18, 16}},
		},
	},
}

var test4 = jsTestCase{
	name: "test arrays and exceptions",
	inputJs: `
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
	want: parsing.ParseResult{
		Identifiers: []parsing.ParsedIdentifier{
			{parsing.Function, "test4", parsing.TextPosition{2, 9}},
			{parsing.Variable, "a", parsing.TextPosition{3, 10}},
			{parsing.Member, "log", parsing.TextPosition{6, 20}},
			{parsing.Member, "log", parsing.TextPosition{8, 20}},
			{parsing.Member, "log", parsing.TextPosition{10, 20}},
			{parsing.Parameter, "e", parsing.TextPosition{12, 13}},
			{parsing.Variable, "f", parsing.TextPosition{13, 12}},
			{parsing.Member, "log", parsing.TextPosition{14, 16}},
			{parsing.Member, "log", parsing.TextPosition{19, 20}},
			{parsing.Member, "log", parsing.TextPosition{22, 20}},
		},
		Literals: []parsing.ParsedLiteral[any]{
			{"Numeric", "float64", 1.0, "1", true, parsing.TextPosition{3, 15}},
			{"Numeric", "float64", 2.0, "2", true, parsing.TextPosition{3, 18}},
			{"Numeric", "float64", 3.0, "3", true, parsing.TextPosition{3, 21}},
			{"Numeric", "float64", 1.0, "1", false, parsing.TextPosition{5, 14}},
			{"Numeric", "float64", 3.0, "3", false, parsing.TextPosition{5, 21}},
			{"Numeric", "float64", 1.0, "1", false, parsing.TextPosition{6, 27}},
			{"Numeric", "float64", 1.0, "1", false, parsing.TextPosition{7, 21}},
			{"Numeric", "float64", 2.0, "2", false, parsing.TextPosition{7, 28}},
			{"Numeric", "float64", 1.0, "1", false, parsing.TextPosition{8, 26}},
			{"Numeric", "float64", 2.0, "2", false, parsing.TextPosition{10, 26}},
			{"String", "string", "abc", `"abc"`, false, parsing.TextPosition{13, 16}},
			{"Numeric", "float64", 0.0, "0", false, parsing.TextPosition{17, 14}},
			{"Numeric", "float64", 1.0, "1", false, parsing.TextPosition{18, 13}},
			{"String", "string", "Hp", `"Hp"`, false, parsing.TextPosition{19, 24}},
			{"String", "string", "Hq", `"Hq"`, false, parsing.TextPosition{22, 24}},
		},
	},
}

var test5 = jsTestCase{
	name: "test class definition",
	inputJs: `
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
	want: parsing.ParseResult{
		Identifiers: []parsing.ParsedIdentifier{
			{parsing.Variable, "Rectangle", parsing.TextPosition{3, 4}},
			{parsing.Parameter, "height", parsing.TextPosition{4, 16}},
			{parsing.Parameter, "width", parsing.TextPosition{4, 24}},
			{parsing.Member, "height", parsing.TextPosition{5, 13}},
			{parsing.Member, "width", parsing.TextPosition{6, 13}},
			{parsing.Member, "log", parsing.TextPosition{9, 8}},
			{parsing.Member, "name", parsing.TextPosition{9, 22}},
			//{Variable, "Rectangle", TextPosition{4, 22}},
			{parsing.Class, "Rectangle2", parsing.TextPosition{13, 18}},
			{parsing.Property, "test", parsing.TextPosition{14, 5}},
			{parsing.Parameter, "height", parsing.TextPosition{15, 16}},
			{parsing.Parameter, "width", parsing.TextPosition{15, 24}},
			{parsing.Member, "height", parsing.TextPosition{16, 13}},
			{parsing.Member, "width", parsing.TextPosition{17, 13}},
			{parsing.Member, "log", parsing.TextPosition{20, 8}},
			{parsing.Member, "name", parsing.TextPosition{20, 22}},
		},
		Literals: []parsing.ParsedLiteral[any]{},
	},
}

func TestParseJS(t *testing.T) {
	const printAllJson = false
	var tests = []jsTestCase{test1, test2, test3, test4, test5}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, parserOutput, err := ParseJS(jsParserPath, "", tt.inputJs)
			if err != nil {
				t.Errorf("ParseJS() error = %v", err)
				println("Parser output:\n", parserOutput)
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

			if t.Failed() || printAllJson {
				println("Raw JSON:\n", parserOutput)
			}

		})
	}
}
