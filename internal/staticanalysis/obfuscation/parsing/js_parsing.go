package parsing

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ParserOutputElement
// Output JSON format of JS parser
type ParserOutputElement struct {
	SymbolType    string         `json:"type"`
	SymbolSubtype string         `json:"subtype"`
	Data          any            `json:"data"`
	Pos           [2]int         `json:"pos"`
	Array         bool           `json:"array"`
	Extra         map[string]any `json:"extra"`
}

// if sourcePath is empty, sourceString will be parsed as JS code
func runParser(parserPath, jsFilePath, jsSource string) (string, error) {
	var out []byte
	var err error
	if len(jsFilePath) > 0 {
		cmd := exec.Command(parserPath, jsFilePath)
		out, err = cmd.Output()
	} else {
		cmd := exec.Command(parserPath)
		pipe, err := cmd.StdinPipe()
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
		//fmt.Printf("Returned output %s\n", string(out))
		return string(out), nil
	} else {
		return "", err
	}

}

// ParseJS
// if sourcePath is empty, sourceString will be parsed as JS code
func ParseJS(parserPath string, filePath string, sourceString string, printJson bool) (*ParseResult, error) {
	parserOutput, err := runParser(parserPath, filePath, sourceString)
	if err != nil {
		return nil, err
	}

	jsonString := parserOutput
	println("Decoding JSON")
	// parse JSON to get results as Go struct
	decoder := json.NewDecoder(strings.NewReader(jsonString))
	var storage []ParserOutputElement
	err = decoder.Decode(&storage)
	if err != nil {
		println("Failed on decoding the following JSON")
		println(jsonString)
		return nil, err
	} else {
		if printJson {
			println(jsonString)
		}
	}

	// convert the elements into more natural data structure
	result := ParseResult{}
	for _, element := range storage {
		switch element.SymbolType {
		case "Identifier":
			symbolSubtype := checkIdentifierType(element.SymbolSubtype)
			if symbolSubtype == Other || symbolSubtype == Unknown {
				break
			}
			result.Identifiers = append(result.Identifiers, ParsedIdentifier{
				Type: checkIdentifierType(element.SymbolSubtype),
				Name: element.Data.(string),
				Pos:  TextPosition{element.Pos[0], element.Pos[1]},
			})
		case "Literal":
			result.Literals = append(result.Literals, ParsedLiteral[any]{
				Type:     fmt.Sprintf("%T", element.Data),
				Value:    element.Data,
				RawValue: element.Extra["raw"].(string),
				InArray:  element.Array,
				Pos:      TextPosition{element.Pos[0], element.Pos[1]},
			})
		default:
			panic(fmt.Errorf("unknown element type for parsed symbol: %s", element.SymbolType))
		}
	}
	return &result, nil
}

func RunExampleParsing(jsParserPath, jsFilePath string) {
	parseResult, err := ParseJS(jsParserPath, jsFilePath, "", true)
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
