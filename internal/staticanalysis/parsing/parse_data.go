package parsing

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/linelengths"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

// Result holds the full set of parsing data for a single file, which may
// contain valid parsing results for multiple languages
type Result map[Language]*Data

// Data (parsing.Data) holds information about source code tokens found
// in a given input text when parsed as particular language.
type Data struct {
	LineLengths    map[int]int
	Identifiers    []token.Identifier
	StringLiterals []token.String
	IntLiterals    []token.Int
	FloatLiterals  []token.Float
	Comments       []token.Comment
}

func (rd Data) String() string {
	parts := []string{
		fmt.Sprintf("line lengths\n%v\n", rd.LineLengths),
		fmt.Sprintf("identifiers\n%v\n", rd.Identifiers),
		fmt.Sprintf("string literals\n%v\n", rd.StringLiterals),
		fmt.Sprintf("integer literals\n%v\n", rd.IntLiterals),
		fmt.Sprintf("float literals\n%v\n", rd.FloatLiterals),
	}
	return strings.Join(parts, "\n-------------------\n")
}

/*
ParseSingle parses the given input source code (either a file or string) using all
supported language parsers and returns a map of language to parsing.Data, which holds
information about the source code tokens found for that language.

Currently, the only supported language is JavaScript, however more language parsers will
be added in the future.

Input can be specified either by file path or by passing the source code string directly.
To parse a file, specify its path using sourceFile; the value of sourceString is ignored.
If sourceFile is empty, then sourceString is parsed directly as code.

If parsing is attempted in a given langauge and fails due to syntax errors, the value
for that language in the returned map is nil, with no other error.

If an internal error occurs during parsing, parsing is interrupted and the error returned.

Note: In JavaScript, there is no distinction between integer and floating point literals;
they are normally both parsed as floating point. This function records a numeric literal
as an integer if it can be converted using strconv.Atoi(), otherwise it is recorded as
floating point.
*/
func ParseSingle(parserConfig ParserConfig, sourceFile string, sourceString string, printDebug bool) (Result, error) {
	parseResult, parserOutput, err := parseJS(parserConfig, sourceFile, sourceString)
	if printDebug {
		fmt.Fprintf(os.Stderr, "\nRaw JSON:\n%s\n", parserOutput)
	}
	if err != nil {
		return nil, err
	}
	if !parseResult.ValidInput {
		return map[Language]*Data{JavaScript: nil}, nil
	}

	lineLengths, err := linelengths.GetLineLengths(sourceFile, sourceString)
	if err != nil {
		return nil, err
	}

	// Initialise with empty slices to avoid null values in JSON
	data := &Data{
		LineLengths:    lineLengths,
		Identifiers:    []token.Identifier{},
		StringLiterals: []token.String{},
		IntLiterals:    []token.Int{},
		FloatLiterals:  []token.Float{},
		Comments:       []token.Comment{},
	}

	for _, d := range parseResult.Literals {
		if d.GoType == "string" {
			data.StringLiterals = append(data.StringLiterals, token.String{Value: d.Value.(string), Raw: d.RawValue})
		} else if d.GoType == "float64" {
			if intValue, err := strconv.ParseInt(d.RawValue, 0, 64); err == nil {
				data.IntLiterals = append(data.IntLiterals, token.Int{Value: intValue, Raw: d.RawValue})
			} else {
				data.FloatLiterals = append(data.FloatLiterals, token.Float{Value: d.Value.(float64), Raw: d.RawValue})
			}
		}
	}

	for _, ident := range parseResult.Identifiers {
		switch ident.Type {
		case token.Function:
			fallthrough
		case token.Class:
			fallthrough
		case token.Parameter:
			fallthrough
		case token.Property:
			fallthrough
		case token.Variable:
			data.Identifiers = append(data.Identifiers, token.Identifier{Name: ident.Name, Type: ident.Type})
		}
	}

	for _, comment := range parseResult.Comments {
		data.Comments = append(data.Comments, token.Comment{Value: comment.Data})
	}

	return map[Language]*Data{JavaScript: data}, nil
}
