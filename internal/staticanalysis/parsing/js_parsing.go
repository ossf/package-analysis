package parsing

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"strings"

	"github.com/ossf/package-analysis/internal/log"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

// parseOutputJSON represents the output JSON format of the JS parser
// it maps filenames to parse data
type parseOutputJSON map[string]parseDataJSON

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
func runParser(parserPath string, input InputStrategy, extraArgs ...string) (string, error) {
	workingDir, err := os.MkdirTemp("", "package-analysis-run-parser-*")
	if err != nil {
		return "", fmt.Errorf("runParser failed to create temp working directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(workingDir); err != nil {
			log.Error("could not remove working directory", "path", workingDir, "error", err)
		}
	}()

	outFilePath := workingDir + string(os.PathSeparator) + "output.json"

	nodeArgs := []string{parserPath, "--output", outFilePath}
	if len(extraArgs) > 0 {
		nodeArgs = append(nodeArgs, extraArgs...)
	}

	cmd := exec.Command("node", nodeArgs...)

	if err := input.SendTo(cmd, workingDir); err != nil {
		return "", fmt.Errorf("runParser failed to prepare parsing input: %w", err)
	}

	if _, err := cmd.Output(); err != nil {
		return "", err
	}

	if output, err := os.ReadFile(outFilePath); err != nil {
		return "", fmt.Errorf("runParser failed to read output file: %w", err)
	} else {
		return string(output), nil
	}
}

func (pd parseDataJSON) process() languageData {
	processed := languageData{
		ValidInput: true,
	}

	// process source code tokens
	for _, t := range pd.Tokens {
		switch t.TokenType {
		case identifier:
			symbolSubtype := token.CheckIdentifierType(t.TokenSubType)
			if symbolSubtype == token.Other || symbolSubtype == token.Unknown {
				break
			}
			processed.Identifiers = append(processed.Identifiers, parsedIdentifier{
				Type: token.CheckIdentifierType(t.TokenSubType),
				Name: t.Data.(string),
				Pos:  t.Pos,
			})
		case literal:
			literal := parsedLiteral[any]{
				Type:     t.TokenSubType,
				GoType:   fmt.Sprintf("%T", t.Data),
				Value:    t.Data,
				RawValue: t.Extra["raw"].(string),
				InArray:  t.Extra["array"] == true,
				Pos:      t.Pos,
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
			processed.Literals = append(processed.Literals, literal)
		case comment:
			processed.Comments = append(processed.Comments, parsedComment{
				Type: t.TokenSubType,
				Data: t.Data.(string),
				Pos:  t.Pos,
			})
		default:
			log.Warn(fmt.Sprintf("parseJS: unrecognised token type %s", t.TokenType))
		}
	}
	// process parser status (info/errors)
	for _, s := range pd.Status {
		status := parserStatus{
			Type:    s.StatusType,
			Name:    s.StatusSubType,
			Message: s.Message,
			Pos:     s.Pos,
		}
		switch s.StatusType {
		case parseInfo:
			processed.Info = append(processed.Info, status)
		case parseError:
			processed.Errors = append(processed.Errors, status)
			if strings.Contains(status.Message, fatalSyntaxErrorMarker) {
				processed.ValidInput = false
			}
		}
	}

	return processed
}

/*
parseJS extracts source code identifiers and string literals from JavaScript code.

parserConfig specifies options relevant to the parser itself, and is produced by InitParser

If internal errors occurred during parsing, then a nil languageResult pointer is returned.
The other two return values are the raw parser output and the error respectively.
Otherwise, the first return value points to the parsing result object while the second
contains the raw JSON output from the parser.
*/
func parseJS(parserConfig ParserConfig, input InputStrategy) (languageResult, string, error) {
	rawOutput, err := runParser(parserConfig.ParserPath, input)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			rawOutput = string(exitErr.Stderr)
		}
		return nil, rawOutput, err
	}

	decoder := json.NewDecoder(strings.NewReader(rawOutput))
	var parseOutput parseOutputJSON
	err = decoder.Decode(&parseOutput)
	if err != nil {
		return nil, rawOutput, err
	}

	// convert the elements into more natural data structure
	result := languageResult{}
	for filename, data := range parseOutput {
		result[filename] = data.process()
	}

	return result, rawOutput, nil
}

func RunExampleParsing(config ParserConfig, input InputStrategy) {
	parseResult, rawOutput, err := parseJS(config, input)

	println("\nRaw JSON:\n", rawOutput)

	if err != nil {
		fmt.Printf("Error: %s\n", err)
		if ee, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Process stderr:\n")
			fmt.Println(string(ee.Stderr))
		}
		return
	}

	for _, data := range parseResult {
		fmt.Printf("%s\n", data.String())
	}
}
