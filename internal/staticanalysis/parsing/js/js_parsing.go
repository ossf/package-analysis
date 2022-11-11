package js

import (
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
)

// parserOutputElement represents the output JSON format of the JS parser
type parserOutputElement struct {
	SymbolType    string         `json:"type"`
	SymbolSubtype string         `json:"subtype"`
	Data          any            `json:"data"`
	Pos           [2]int         `json:"pos"`
	Array         bool           `json:"array"`
	Extra         map[string]any `json:"extra"`
}

// runParser handles calling a parser program and provide the specified Javascript source to it, either
// by filename (jsFilePath) or piping jsSource to the program's stdin
// sourcePath is empty, sourceString will be parsed as JS code
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

// ParseJS extracts source code identifiers and string literals from JavaScript code
// if sourcePath is empty, sourceString will be parsed as JS code
func ParseJS(parserPath string, filePath string, sourceString string) (*parsing.ParseResult, string, error) {
	parserOutput, err := runParser(parserPath, filePath, sourceString)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			parserOutput = string(exitErr.Stderr)
		}
		return nil, parserOutput, err
	}

	// parse JSON to get results as Go struct
	decoder := json.NewDecoder(strings.NewReader(parserOutput))
	var storage []parserOutputElement
	err = decoder.Decode(&storage)
	if err != nil {
		return nil, parserOutput, err
	}

	// convert the elements into more natural data structure
	result := parsing.ParseResult{}
	for _, element := range storage {
		switch element.SymbolType {
		case "Identifier":
			symbolSubtype := parsing.CheckIdentifierType(element.SymbolSubtype)
			if symbolSubtype == parsing.Other || symbolSubtype == parsing.Unknown {
				break
			}
			result.Identifiers = append(result.Identifiers, parsing.ParsedIdentifier{
				Type: parsing.CheckIdentifierType(element.SymbolSubtype),
				Name: element.Data.(string),
				Pos:  element.Pos,
			})
		case "Literal":
			result.Literals = append(result.Literals, parsing.ParsedLiteral[any]{
				Type:     element.SymbolSubtype,
				GoType:   fmt.Sprintf("%T", element.Data),
				Value:    element.Data,
				RawValue: element.Extra["raw"].(string),
				InArray:  element.Array,
				Pos:      element.Pos,
			})
		default:
			panic(fmt.Errorf("unknown element type for parsed symbol: %s", element.SymbolType))
		}
	}
	return &result, parserOutput, nil
}

func RunExampleParsing(jsParserPath, jsFilePath string, jsSourceString string) {
	parseResult, jsonString, err := ParseJS(jsParserPath, jsFilePath, jsSourceString)
	if jsonString != "" {
		println("\nRaw JSON:\n", jsonString)
	}
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
		fmt.Println(identifier)
	}
	println()
	println("== Parsed Literals ==")
	for _, literal := range parseResult.Literals {
		fmt.Println(literal)
	}
}
