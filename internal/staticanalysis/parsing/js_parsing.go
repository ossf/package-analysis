package parsing

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os/exec"
	"strings"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

// parseOutput represents the output JSON format of the JS parser
type parseDataJSON struct {
	Tokens []parserTokenJSON  `json:"tokens"`
	Status []parserStatusJSON `json:"status"`
}

type parserTokenJSON struct {
	TokenType    tokenType      `json:"type"`
	TokenSubType string         `json:"subtype"`
	Data         any            `json:"data"`
	Pos          [2]int         `json:"pos"`
	Extra        map[string]any `json:"extra"`
}

type parserStatusJSON struct {
	StatusType    statusType `json:"type"`
	StatusSubType string     `json:"subtype"`
	Message       string     `json:"data"`
	Pos           [2]int     `json:"pos"`
}

// fatalSyntaxErrorMarker is used by the parser to signal that it is unable to
// parse a file completely due to syntax errors that cannot be recovered from.
const fatalSyntaxErrorMarker = "FATAL SYNTAX ERROR"

/*
runParser handles calling the parser program and provide the specified Javascript source to it,
either by filename (jsFilePath) or piping jsSource to the program's stdin.
If sourcePath is empty, sourceString will be parsed as JS code
*/
func runParser(parserPath, jsFilePath, jsSource string, extraArgs ...string) (string, error) {
	nodeArgs := []string{parserPath}
	if len(jsFilePath) > 0 {
		nodeArgs = append(nodeArgs, jsFilePath)
	}
	if len(extraArgs) > 0 {
		nodeArgs = append(nodeArgs, extraArgs...)
	}

	cmd := exec.Command("node", nodeArgs...)

	if len(jsFilePath) == 0 {
		// create a pipe to send the source code to the parser via stdin
		pipe, pipeErr := cmd.StdinPipe()
		if pipeErr != nil {
			return "", fmt.Errorf("runParser failed to create pipe: %v", pipeErr)
		}

		if _, pipeErr = pipe.Write([]byte(jsSource)); pipeErr != nil {
			return "", fmt.Errorf("runParser failed to write source string to pipe: %w", pipeErr)
		}

		if pipeErr = pipe.Close(); pipeErr != nil {
			return "", fmt.Errorf("runParser failed to close pipe: %w", pipeErr)
		}
	}

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}

/*
parseJS extracts source code identifiers and string literals from JavaScript code.
If sourcePath is empty, sourceString will be parsed as JS code.

parserConfig specifies options relevant to the parser itself, and is produced by InitParser

If internal errors occurred during parsing, then a nil parseOutput pointer is returned.
The other two return values are the raw parser output and the error object respectively.
Otherwise, the first return value points to the results of parsing while the second
contains the raw JSON output from the parser.
*/
func parseJS(parserConfig ParserConfig, filePath string, sourceString string) (*parseOutput, string, error) {
	parserOutput, err := runParser(parserConfig.ParserPath, filePath, sourceString)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			parserOutput = string(exitErr.Stderr)
		}
		return nil, parserOutput, err
	}

	// parse JSON to get results as Go struct
	decoder := json.NewDecoder(strings.NewReader(parserOutput))
	var parserData parseDataJSON
	err = decoder.Decode(&parserData)
	if err != nil {
		return nil, parserOutput, err
	}

	result := &parseOutput{
		ValidInput: true,
	}

	// convert the elements into more natural data structure
	for _, element := range parserData.Tokens {
		switch element.TokenType {
		case identifier:
			symbolSubtype := token.CheckIdentifierType(element.TokenSubType)
			if symbolSubtype == token.Other || symbolSubtype == token.Unknown {
				break
			}
			result.Identifiers = append(result.Identifiers, parsedIdentifier{
				Type: token.CheckIdentifierType(element.TokenSubType),
				Name: element.Data.(string),
				Pos:  element.Pos,
			})
		case literal:
			literal := parsedLiteral[any]{
				Type:     element.TokenSubType,
				GoType:   fmt.Sprintf("%T", element.Data),
				Value:    element.Data,
				RawValue: element.Extra["raw"].(string),
				InArray:  element.Extra["array"] == true,
				Pos:      element.Pos,
			}
			// check for BigInteger types which have to be represented as strings in JSON
			if literal.Type == "Numeric" && literal.GoType == "string" {
				if intAsString, ok := literal.Value.(string); ok {
					var bigInt big.Int
					if _, valid := bigInt.SetString(intAsString, 0); valid {
						literal.Value = &bigInt
						literal.GoType = fmt.Sprintf("%T", bigInt)
					}
				}
			}
			result.Literals = append(result.Literals, literal)
		case comment:
			result.Comments = append(result.Comments, parsedComment{
				Type: element.TokenSubType,
				Data: element.Data.(string),
				Pos:  element.Pos,
			})
		default:
			log.Warn(fmt.Sprintf("parseJS: unrecognised token type %s", element.TokenType))
		}
	}
	for _, element := range parserData.Status {
		status := parserStatus{
			Type:    element.StatusType,
			Name:    element.StatusSubType,
			Message: element.Message,
			Pos:     element.Pos,
		}
		switch element.StatusType {
		case parseInfo:
			result.Info = append(result.Info, status)
		case parseError:
			result.Errors = append(result.Errors, status)
			if strings.Contains(status.Message, fatalSyntaxErrorMarker) {
				result.ValidInput = false
			}
		}
	}

	return result, parserOutput, nil
}

func RunExampleParsing(config ParserConfig, jsFilePath string, jsSourceString string) {
	parseResult, parserOutput, err := parseJS(config, jsFilePath, jsSourceString)

	println("\nRaw JSON:\n", parserOutput)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		if ee, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Process stderr:\n")
			fmt.Println(string(ee.Stderr))
		}
		return
	}

	fmt.Printf("%s\n", parseResult.String())
}
