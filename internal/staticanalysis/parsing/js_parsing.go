package parsing

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/externalcmd"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis/token"
)

// parseOutputJSON represents the output JSON format of the JS parser
// it maps filenames to parse data.
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

// parserArgsHandler specifies how to pass CLI args for the parser to externalcmd.Input.
type parserArgsHandler struct{}

func (h parserArgsHandler) SingleFileArg(filePath string) []string {
	return []string{"--file", filePath}
}

func (h parserArgsHandler) FileListArg(fileListPath string) []string {
	return []string{"--batch", fileListPath}
}

func (h parserArgsHandler) ReadStdinArg() []string {
	return []string{} // no argument needed
}

/*
runParser handles calling the parser program and provide the specified Javascript source to it,
either by filename (jsFilePath) or piping jsSource to the program's stdin.

If sourcePath is empty, sourceString will be parsed as JS code.
*/
func runParser(ctx context.Context, parserPath string, input externalcmd.Input, extraArgs ...string) (string, error) {
	workingDir, err := os.MkdirTemp("", "package-analysis-run-parser-*")
	if err != nil {
		return "", fmt.Errorf("runParser failed to create temp working directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(workingDir); err != nil {
			slog.ErrorContext(ctx, "could not remove working directory", "path", workingDir, "error", err)
		}
	}()

	outFilePath := filepath.Join(workingDir, "output.json")

	nodeArgs := []string{parserPath, "--output", outFilePath}
	if len(extraArgs) > 0 {
		nodeArgs = append(nodeArgs, extraArgs...)
	}

	cmd := exec.CommandContext(ctx, "node", nodeArgs...)

	if err := input.SendTo(cmd, parserArgsHandler{}, workingDir); err != nil {
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

func (pd parseDataJSON) process(ctx context.Context) singleParseData {
	processed := singleParseData{
		ValidInput: true,
	}

	// process source code tokens
	for _, t := range pd.Tokens {
		switch t.TokenType {
		case identifier:
			symbolSubtype := token.ParseIdentifierType(t.TokenSubType)
			if symbolSubtype == token.Other || symbolSubtype == token.Unknown {
				break
			}
			processed.Identifiers = append(processed.Identifiers, parsedIdentifier{
				Type: token.ParseIdentifierType(t.TokenSubType),
				Name: t.Data.(string),
				Pos:  t.Pos,
			})
		case literal:
			literal := parsedLiteral[any]{
				Type:    t.TokenSubType,
				GoType:  fmt.Sprintf("%T", t.Data),
				Value:   t.Data,
				InArray: t.Extra["array"] == true,
				Pos:     t.Pos,
			}

			// Since t.Extra is a map[string]any, t.Extra["raw"].(string) will panic
			// if "raw" is not present in the map, due to nil -> string conversion.
			// Therefore we need the conditional type assertion as below.
			if rawValue, ok := t.Extra["raw"].(string); ok {
				literal.RawValue = rawValue
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
			slog.WarnContext(ctx, fmt.Sprintf("parseJS: unrecognised token type %s", t.TokenType))
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

If internal errors occurred during parsing, then a nil map is returned.
The other two return values are the raw parser output and the error respectively.
Otherwise, the first return value points to the parsing result object while the second
contains the raw JSON output from the parser.
*/
func parseJS(ctx context.Context, parserConfig ParserConfig, input externalcmd.Input) (map[string]singleParseData, string, error) {
	rawOutput, err := runParser(ctx, parserConfig.ParserPath, input)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			rawOutput = string(exitErr.Stderr)
		}
		return nil, rawOutput, err
	}

	decoder := json.NewDecoder(strings.NewReader(rawOutput))

	var parseOutput parseOutputJSON
	if err := decoder.Decode(&parseOutput); err != nil {
		return nil, rawOutput, err
	}

	// convert the elements into more natural data structure
	result := map[string]singleParseData{}
	for filename, data := range parseOutput {
		result[filename] = data.process(ctx)
	}

	return result, rawOutput, nil
}

func RunExampleParsing(ctx context.Context, config ParserConfig, input externalcmd.Input) {
	parseResult, rawOutput, err := parseJS(ctx, config, input)

	fmt.Println("\nRaw JSON:\n", rawOutput)

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
