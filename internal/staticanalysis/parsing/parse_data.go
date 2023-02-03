package parsing

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

// Result holds the full set of parsing data for a single file, which may
// contain valid parsing results for multiple languages
type Result map[Language]*Data

// Data (parsing.Data) holds information about source code tokens found
// in a given input text when parsed as particular language.
type Data struct {
	SourceLanguage Language
	Identifiers    []token.Identifier
	StringLiterals []token.String
	IntLiterals    []token.Int
	FloatLiterals  []token.Float
	Comments       []token.Comment
}

func (pd Data) String() string {
	parts := []string{
		fmt.Sprintf("source language: %s\n", pd.SourceLanguage),
		fmt.Sprintf("identifiers\n%v\n", pd.Identifiers),
		fmt.Sprintf("string literals\n%v\n", pd.StringLiterals),
		fmt.Sprintf("integer literals\n%v\n", pd.IntLiterals),
		fmt.Sprintf("float literals\n%v\n", pd.FloatLiterals),
		fmt.Sprintf("comments\n%v\n", pd.Comments),
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
If sourceFile is empty, then jsSourceString is parsed directly as code.

If parsing is attempted in a given langauge and fails due to syntax errors, the value
for that language in the returned map is nil, with no other error.

If an internal error occurs during parsing, parsing is interrupted and the error returned.

Note: In JavaScript, there is no distinction between integer and floating point literals;
they are normally both parsed as floating point. This function records a numeric literal
as an integer if it can be converted using strconv.Atoi(), otherwise it is recorded as
floating point.
*/
func ParseSingle(parserConfig ParserConfig, sourceFile string, sourceString string, printJSON bool) (Result, error) {
	parseResult, parserOutput, err := ParseJS(parserConfig, sourceFile, sourceString)
	if printJSON {
		fmt.Fprintf(os.Stderr, "\nRaw JSON:\n%s\n", parserOutput)
	}
	if err != nil {
		return nil, err
	}

	if !parseResult.ValidInput {
		return map[Language]*Data{JavaScript: nil}, nil
	}

	// Initialise with empty slices to avoid null values in JSON
	jsData := &Data{
		SourceLanguage: JavaScript,
		Identifiers:    []token.Identifier{},
		StringLiterals: []token.String{},
		IntLiterals:    []token.Int{},
		FloatLiterals:  []token.Float{},
		Comments:       []token.Comment{},
	}

	for _, d := range parseResult.Literals {
		if d.GoType == "string" {
			jsData.StringLiterals = append(jsData.StringLiterals, token.String{Value: d.Value.(string), Raw: d.RawValue})
		} else if d.GoType == "float64" {
			if intValue, err := strconv.ParseInt(d.RawValue, 0, 64); err == nil {
				jsData.IntLiterals = append(jsData.IntLiterals, token.Int{Value: intValue, Raw: d.RawValue})
			} else {
				jsData.FloatLiterals = append(jsData.FloatLiterals, token.Float{Value: d.Value.(float64), Raw: d.RawValue})
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
			jsData.Identifiers = append(jsData.Identifiers, token.Identifier{Name: ident.Name, Type: ident.Type})
		}
	}

	for _, comment := range parseResult.Comments {
		jsData.Comments = append(jsData.Comments, token.Comment{Value: comment.Data})
	}

	return map[Language]*Data{JavaScript: jsData}, nil
}
