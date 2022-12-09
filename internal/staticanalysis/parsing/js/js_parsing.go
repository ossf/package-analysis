package js

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
)

// parserOutputElement represents the output JSON format of the JS parser
type parserOutputElement struct {
	SymbolType    parsing.SymbolType `json:"type"`
	SymbolSubtype string             `json:"subtype"`
	Data          any                `json:"data"`
	Pos           [2]int             `json:"pos"`
	Extra         map[string]any     `json:"extra"`
}

/*
syntaxErrorExitCode is the exit code that the parser will return if it encounters a
syntax error while parsing the input. This also ends up being the signal of whether a given
input is JavaScript or not - without an external tool that detects file types, it's hard
to tell between 'JavaScript with a few errors' and 'a totally non-JavaScript file'.
*/
const syntaxErrorExitCode = 33

/*
runParser handles calling the parser program and provide the specified Javascript source to it,
either by filename (jsFilePath) or piping jsSource to the program's stdin.
If sourcePath is empty, sourceString will be parsed as JS code
*/
func runParser(parserPath, jsFilePath, jsSource string) (string, error) {
	var out []byte
	var err error
	if len(jsFilePath) > 0 {
		cmd := exec.Command(parserPath, jsFilePath)
		out, err = cmd.Output()
	} else {
		cmd := exec.Command(parserPath)
		var pipe io.WriteCloser
		pipe, err = cmd.StdinPipe()
		if err == nil {
			var bytesWritten int
			//fmt.Printf("Writing %s\n", jsSource)
			bytesWritten, err = pipe.Write([]byte(jsSource))
			if err == nil && bytesWritten != len(jsSource) {
				// couldn't write all data
				err = fmt.Errorf("failed to pipe source string to parser (%d of %d bytes written)",
					bytesWritten, len(jsSource))
			}
			//fmt.Printf("Wrote %d bytes\n", bytesWritten)
			err = pipe.Close()
		}
		if err == nil {
			out, err = cmd.Output()
		}
	}

	if err == nil {
		return string(out), nil
	}

	return "", err
}

/*
ParseJS extracts source code identifiers and string literals from JavaScript code.
If sourcePath is empty, sourceString will be parsed as JS code.

parserConfig specifies options relevant to the parser itself, and is produced by InitParser

If the input contains a syntax error (which could mean it's not actually JavaScript),
then a pointer to parsing.InvalidInput is returned.
*/
func ParseJS(parserConfig ParserConfig, filePath string, sourceString string) (result parsing.ParseResult, parserOutput string, err error) {
	if err != nil {
		return
	}

	parserOutput, err = runParser(parserConfig.ParserPath, filePath, sourceString)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == syntaxErrorExitCode {
				return parsing.InvalidInput, "", nil
			}
			parserOutput = string(exitErr.Stderr)
		}
		return
	}

	// parse JSON to get results as Go struct
	decoder := json.NewDecoder(strings.NewReader(parserOutput))
	var storage []parserOutputElement
	err = decoder.Decode(&storage)
	if err != nil {
		return
	}

	result.ValidInput = true

	// convert the elements into more natural data structure
	for _, element := range storage {
		switch element.SymbolType {
		case parsing.Identifier:
			symbolSubtype := parsing.CheckIdentifierType(element.SymbolSubtype)
			if symbolSubtype == parsing.Other || symbolSubtype == parsing.Unknown {
				break
			}
			result.Identifiers = append(result.Identifiers, parsing.ParsedIdentifier{
				Type: parsing.CheckIdentifierType(element.SymbolSubtype),
				Name: element.Data.(string),
				Pos:  element.Pos,
			})
		case parsing.Literal:
			result.Literals = append(result.Literals, parsing.ParsedLiteral[any]{
				Type:     element.SymbolSubtype,
				GoType:   fmt.Sprintf("%T", element.Data),
				Value:    element.Data,
				RawValue: element.Extra["raw"].(string),
				InArray:  element.Extra["array"] == true,
				Pos:      element.Pos,
			})
		case parsing.Comment:
			result.Comments = append(result.Comments, parsing.ParsedComment{
				Type: element.SymbolSubtype,
				Data: element.Data.(string),
				Pos:  element.Pos,
			})
		case parsing.Info:
			fallthrough
		case parsing.Error:
			// ignore for now
		default:
			log.Warn(fmt.Sprintf("ParseJS: unrecognised symbol type %s", element.SymbolType))
		}
	}
	return
}

func RunExampleParsing(config ParserConfig, jsFilePath string, jsSourceString string) {
	parseResult, parserOutput, err := ParseJS(config, jsFilePath, jsSourceString)

	println("\nRaw JSON:\n", parserOutput)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		if ee, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Process stderr:\n")
			fmt.Println(string(ee.Stderr))
		}
		return
	} else {
		fmt.Println("Completed without errors")
	}
	println()
	println("== Parsed Identifiers ==")
	for _, identifier := range parseResult.Identifiers {
		fmt.Printf("%v\n", identifier)
	}
	println()
	println("== Parsed Literals ==")
	for _, literal := range parseResult.Literals {
		fmt.Printf("%v\n", literal)
	}

	println()
	println("== Parsed Comments ==")
	for _, comment := range parseResult.Comments {
		fmt.Printf("%v\n", comment)
	}

}
