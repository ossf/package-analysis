package parsing

import (
	"reflect"
	"testing"
)

const jsParserPath = "../parsing/babel-parser.js"

type jsTestCase struct {
	name      string
	inputJs   string
	want      ParseResult
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
	want: ParseResult{
		Identifiers: []ParsedIdentifier{
			{Function, "test", TextPosition{2, 9}},
			{Variable, "mystring1", TextPosition{3, 8}},
			{Variable, "mystring2", TextPosition{4, 8}},
			{Variable, "mystring3", TextPosition{5, 8}},
			{Variable, "mystring4", TextPosition{6, 8}},
			{Variable, "mystring5", TextPosition{7, 8}},
			{Variable, "mystring6", TextPosition{8, 8}},
			{Variable, "mystring7", TextPosition{9, 8}},
			{Variable, "mystring8", TextPosition{10, 8}},
			{Variable, "mystring9", TextPosition{11, 8}},
			{Variable, "mystring10", TextPosition{12, 8}},
			{Variable, "mystring11", TextPosition{13, 5}},
			{Variable, "mystring12", TextPosition{15, 5}},
		},
		Literals: []ParsedLiteral[any]{
			{"String", "string", "hello1", `"hello1"`, false, TextPosition{3, 20}},
			{"String", "string", "hello2", `'hello2'`, false, TextPosition{4, 20}},
			{"String", "string", "hello'3'", `"hello'3'"`, false, TextPosition{5, 20}},
			{"String", "string", "hello\"4\"", `'hello"4"'`, false, TextPosition{6, 20}},
			{"String", "string", "hello\"5\"", `"hello\"5\""`, false, TextPosition{7, 20}},
			{"String", "string", "hello'6'", `"hello\'6\'"`, false, TextPosition{8, 20}},
			{"String", "string", "hello'7'", `'hello\'7\''`, false, TextPosition{9, 20}},
			{"String", "string", "hello", `"hello"`, false, TextPosition{10, 20}},
			{"String", "string", "8", `"8"`, false, TextPosition{10, 30}},
			{"StringTemplate", "string", "hello9", `hello9`, false, TextPosition{11, 21}},
			{"StringTemplate", "string", "hello\"'", `hello"'`, false, TextPosition{12, 22}},
			{"StringTemplate", "string", "\"'", `"'`, false, TextPosition{12, 34}},
			{"Numeric", "float64", 10.0, "10", false, TextPosition{12, 31}},
			{"StringTemplate", "string", "hello\n//\"'11\"'", `hello` + "\n" + `//"'11"'`, false, TextPosition{13, 19}},
			{"StringTemplate", "string", "hello\"'", `hello"'`, false, TextPosition{15, 19}},
			{"StringTemplate", "string", "\"'", `"'`, false, TextPosition{15, 38}},
			{"Numeric", "float64", 5.6, "5.6", false, TextPosition{15, 28}},
			{"Numeric", "float64", 6.4, "6.4", false, TextPosition{15, 34}},
		},
	},
}

var test2 = jsTestCase{
	name: "test function parameters",
	inputJs: `
function test2(param1, param2, param3 = "ahd") {
	return param1 + param2 + param3;
}`,
	want: ParseResult{
		Identifiers: []ParsedIdentifier{
			{Function, "test2", TextPosition{2, 9}},
			{Parameter, "param1", TextPosition{2, 15}},
			{Parameter, "param2", TextPosition{2, 23}},
			{Parameter, "param3", TextPosition{2, 31}},
		},
		Literals: []ParsedLiteral[any]{
			{"String", "string", "ahd", `"ahd"`, false, TextPosition{2, 40}},
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
	want: ParseResult{
		Identifiers: []ParsedIdentifier{
			{Function, "test3", TextPosition{2, 9}},
			{Parameter, "a", TextPosition{2, 15}},
			{Parameter, "b", TextPosition{2, 18}},
			{Parameter, "c", TextPosition{2, 21}},
			{Variable, "i", TextPosition{3, 13}},
			{StatementLabel, "outer", TextPosition{4, 0}},
			{Variable, "j", TextPosition{5, 17}},
			{Variable, "k", TextPosition{6, 21}},
			{Member, "log", TextPosition{16, 16}},
			{Member, "log", TextPosition{18, 12}},
		},
		Literals: []ParsedLiteral[any]{
			{"Numeric", "float64", 1.0, "1", false, TextPosition{5, 21}},
			{"Numeric", "float64", 3.0, "3", false, TextPosition{5, 28}},
			{"Numeric", "float64", 10.0, "10", false, TextPosition{6, 36}},
			{"Numeric", "float64", 2.0, "2", false, TextPosition{7, 26}},
			{"Numeric", "float64", 32.0, "32", false, TextPosition{13, 16}},
			{"Numeric", "float64", 0.0, "0", false, TextPosition{13, 23}},
			{"String", "string", "here", `"here"`, false, TextPosition{16, 20}},
			{"String", "string", "End", `"End"`, false, TextPosition{18, 16}},
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
	want: ParseResult{
		Identifiers: []ParsedIdentifier{
			{Function, "test4", TextPosition{2, 9}},
			{Variable, "a", TextPosition{3, 10}},
			{Member, "log", TextPosition{6, 20}},
			{Member, "log", TextPosition{8, 20}},
			{Member, "log", TextPosition{10, 20}},
			{Parameter, "e", TextPosition{12, 13}},
			{Variable, "f", TextPosition{13, 12}},
			{Member, "log", TextPosition{14, 16}},
			{Member, "log", TextPosition{19, 20}},
			{Member, "log", TextPosition{22, 20}},
		},
		Literals: []ParsedLiteral[any]{
			{"Numeric", "float64", 1.0, "1", true, TextPosition{3, 15}},
			{"Numeric", "float64", 2.0, "2", true, TextPosition{3, 18}},
			{"Numeric", "float64", 3.0, "3", true, TextPosition{3, 21}},
			{"Numeric", "float64", 1.0, "1", false, TextPosition{5, 14}},
			{"Numeric", "float64", 3.0, "3", false, TextPosition{5, 21}},
			{"Numeric", "float64", 1.0, "1", false, TextPosition{6, 27}},
			{"Numeric", "float64", 1.0, "1", false, TextPosition{7, 21}},
			{"Numeric", "float64", 2.0, "2", false, TextPosition{7, 28}},
			{"Numeric", "float64", 1.0, "1", false, TextPosition{8, 26}},
			{"Numeric", "float64", 2.0, "2", false, TextPosition{10, 26}},
			{"String", "string", "abc", `"abc"`, false, TextPosition{13, 16}},
			{"Numeric", "float64", 0.0, "0", false, TextPosition{17, 14}},
			{"Numeric", "float64", 1.0, "1", false, TextPosition{18, 13}},
			{"String", "string", "Hp", `"Hp"`, false, TextPosition{19, 24}},
			{"String", "string", "Hq", `"Hq"`, false, TextPosition{22, 24}},
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
	want: ParseResult{
		Identifiers: []ParsedIdentifier{
			{Variable, "Rectangle", TextPosition{3, 4}},
			{Parameter, "height", TextPosition{4, 16}},
			{Parameter, "width", TextPosition{4, 24}},
			{Member, "height", TextPosition{5, 13}},
			{Member, "width", TextPosition{6, 13}},
			{Member, "log", TextPosition{9, 8}},
			{Member, "name", TextPosition{9, 22}},
			//{Variable, "Rectangle", TextPosition{4, 22}},
			{Class, "Rectangle2", TextPosition{13, 18}},
			{Property, "test", TextPosition{14, 5}},
			{Parameter, "height", TextPosition{15, 16}},
			{Parameter, "width", TextPosition{15, 24}},
			{Member, "height", TextPosition{16, 13}},
			{Member, "width", TextPosition{17, 13}},
			{Member, "log", TextPosition{20, 8}},
			{Member, "name", TextPosition{20, 22}},
		},
		Literals: []ParsedLiteral[any]{},
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
