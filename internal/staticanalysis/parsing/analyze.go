package parsing

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

func (p PackageResult) processAndAdd(language Language, result languageResult) {
	for fileName, fileResult := range result {
		if p[fileName] == nil {
			p[fileName] = FileResult{}
		}

		if !fileResult.ValidInput {
			p[fileName][language] = nil
			continue
		}

		// Initialise with empty slices to avoid null values in JSON
		data := &SingleResult{
			Identifiers:    []token.Identifier{},
			StringLiterals: []token.String{},
			IntLiterals:    []token.Int{},
			FloatLiterals:  []token.Float{},
			Comments:       []token.Comment{},
		}

		for _, d := range fileResult.Literals {
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

		for _, ident := range fileResult.Identifiers {
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

		for _, c := range fileResult.Comments {
			data.Comments = append(data.Comments, token.Comment{Value: c.Data})
		}

		p[fileName][language] = data
	}
}

/*
Analyze (parsing.Analyze) parses the specified list of files using all supported parsers
and returns a map of file path to parsing.FileResult, which in turn is a map of language
to parsing.SingleResult. Each parsing.SingleResult holds information about source code
tokens found for that language.

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
func Analyze(parserConfig ParserConfig, input InputSpec, printDebug bool) (PackageResult, error) {
	resultsByLanguage := map[Language]languageResult{}

	// JavaScript parsing
	jsResults, rawOutput, err := parseJS(parserConfig, input)
	if printDebug {
		fmt.Fprintf(os.Stderr, "\nRaw JSON:\n%s\n", rawOutput)
	}
	if err != nil {
		return nil, err
	}

	packageResult := PackageResult{}
	resultsByLanguage[JavaScript] = jsResults
	for language, result := range resultsByLanguage {
		packageResult.processAndAdd(language, result)
	}

	return packageResult, nil
}
