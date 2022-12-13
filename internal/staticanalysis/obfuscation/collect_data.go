package obfuscation

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ossf/package-analysis/internal/staticanalysis/linelengths"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing/js"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

/*
CollectData parses the given JavaScript input (source file or raw source string) and records raw data
which is needed for processing by ComputeSignals. To parse a file, specify its path using jsSourceFile.
In this case, the value of jsSourceString is ignored. If jsSourceFile is empty, then jsSourceString
is parsed directly as JavaScript code. The input is assumed to be valid JavaScript source.

If a syntax error is found, a nil pointer is returned with no error. This indicates that
the file may not be JavaScript and could be parsed using other methods.

In Javascript, there is little distinction between integer and floating point literals - they are
all parsed as floating point. This function will record a numeric literal as an integer if it can be
converted to an integer using strconv.Atoi(), otherwise it will be recorded as a floating point literal.

Current data collected:
  - list of identifiers (e.g. variable, function, and class names, loop labels)
  - lists of string, integer and floating point literals

TODO planned data
  - recording of arrays of either string literals or numeric data
*/
func CollectData(parserConfig js.ParserConfig, jsSourceFile string, jsSourceString string, printDebug bool) (*RawData, error) {
	parseResult, parserOutput, err := js.ParseJS(parserConfig, jsSourceFile, jsSourceString)
	if printDebug {
		fmt.Fprintf(os.Stderr, "\nRaw JSON:\n%s\n", parserOutput)
	}
	if err != nil {
		return nil, err
	}
	if !parseResult.ValidInput {
		return nil, nil
	}

	lineLengths, err := linelengths.GetLineLengths(jsSourceFile, jsSourceString)
	if err != nil {
		return nil, err
	}

	// Initialise with empty slices to avoid null values in JSON
	data := RawData{
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

	return &data, nil
}
