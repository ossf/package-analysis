package obfuscation

import (
	"fmt"
	"strconv"

	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing/js"
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
		println("\nRaw JSON:\n", parserOutput)
	}
	if err != nil {
		fmt.Printf("Error occured while reading %s: %v\n", jsSourceFile, err)
		return nil, err
	}
	if !parseResult.ValidInput {
		return nil, nil
	}

	// don't want null values in JSON
	data := RawData{
		Identifiers:    make([]string, 0),
		StringLiterals: make([]string, 0),
		IntLiterals:    make([]int, 0),
		FloatLiterals:  make([]float64, 0),
		Comments:       make([]string, 0),
	}

	for _, d := range parseResult.Literals {
		if d.GoType == "string" {
			data.StringLiterals = append(data.StringLiterals, d.Value.(string))
		} else if d.GoType == "float64" {
			// check if it can be an int
			if intValue, err := strconv.Atoi(d.RawValue); err == nil {
				data.IntLiterals = append(data.IntLiterals, intValue)
			} else {
				data.FloatLiterals = append(data.FloatLiterals, d.Value.(float64))
			}
		}
	}

	for _, ident := range parseResult.Identifiers {
		switch ident.Type {
		case parsing.Function:
			fallthrough
		case parsing.Class:
			fallthrough
		case parsing.Parameter:
			fallthrough
		case parsing.Property:
			fallthrough
		case parsing.Variable:
			data.Identifiers = append(data.Identifiers, ident.Name)
		}
	}

	for _, comment := range parseResult.Comments {
		data.Comments = append(data.Comments, comment.Data)
	}

	return &data, nil
}
