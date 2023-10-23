package parsing

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/ossf/package-analysis/internal/staticanalysis/externalcmd"
	"github.com/ossf/package-analysis/internal/staticanalysis/signals/stringentropy"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis/token"
)

func processJsData(fileData singleParseData) SingleResult {
	result := SingleResult{
		Language: NoLanguage,
		// Initialise with empty slices to avoid null values in JSON
		Identifiers:    []token.Identifier{},
		StringLiterals: []token.String{},
		IntLiterals:    []token.Int{},
		FloatLiterals:  []token.Float{},
		Comments:       []token.Comment{},
	}

	if !fileData.ValidInput {
		return result
	}

	// JavaScript is the only currently supported / valid language
	result.Language = JavaScript

	for _, d := range fileData.Literals {
		if d.GoType == "string" {
			result.StringLiterals = append(result.StringLiterals, token.String{Value: d.Value.(string), Raw: d.RawValue})
		} else if d.GoType == "float64" {
			if intValue, err := strconv.ParseInt(d.RawValue, 0, 64); err == nil {
				result.IntLiterals = append(result.IntLiterals, token.Int{Value: intValue, Raw: d.RawValue})
			} else {
				result.FloatLiterals = append(result.FloatLiterals, token.Float{Value: d.Value.(float64), Raw: d.RawValue})
			}
		}
	}

	for _, ident := range fileData.Identifiers {
		switch ident.Type {
		// token.Member is not included as it's too noisy (e.g. console.log)
		case token.Function, token.Class, token.Parameter, token.Property, token.Variable:
			result.Identifiers = append(result.Identifiers, token.Identifier{Name: ident.Name, Type: ident.Type})
		}
	}

	for _, c := range fileData.Comments {
		result.Comments = append(result.Comments, token.Comment{Text: c.Data})
	}
	return result
}

// computeCharacterDistributions estimates the probabilities for characters in
// identifiers and string literals respectively, by aggregating character counts
// across all symbols of each type in the package.
func computeCharacterDistributions(parseResults map[string]SingleResult) (map[rune]float64, map[rune]float64) {
	var identifiers []string
	var strings []string

	for _, r := range parseResults {
		for _, str := range r.StringLiterals {
			strings = append(strings, str.Value)
		}
		for _, ident := range r.Identifiers {
			identifiers = append(identifiers, ident.Name)
		}
	}

	return stringentropy.CharacterProbabilities(identifiers), stringentropy.CharacterProbabilities(strings)
}

/*
Analyze (parsing.Analyze) parses the specified list of files using all supported parsers
and returns a map of filename to slice of parsing.SingleResult. Each slice holds information
about source code tokens found for that file for each supported langauge parser.

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
func Analyze(ctx context.Context, parserConfig ParserConfig, input externalcmd.Input, printDebug bool) (map[string]SingleResult, error) {
	// JavaScript parsing
	jsResults, rawOutput, err := parseJS(ctx, parserConfig, input)
	if printDebug {
		fmt.Fprintf(os.Stderr, "\nRaw JSON:\n%s\n", rawOutput)
	}
	if err != nil {
		return nil, err
	}

	resultsByFile := make(map[string]SingleResult)
	for filename, jsData := range jsResults {
		resultsByFile[filename] = processJsData(jsData)
	}

	// TODO replace this with a global count across many packages from an ecosystem.
	//  If more languages are added before this is done, the function below should be
	//  modified to compute a separate distribution for identifiers each language.
	identifierProbs, stringProbs := computeCharacterDistributions(resultsByFile)

	// populate entropy values for identifiers and string literals.
	for _, r := range resultsByFile {
		for i := range r.Identifiers {
			r.Identifiers[i].ComputeEntropy(identifierProbs)
		}
		for i := range r.StringLiterals {
			r.StringLiterals[i].ComputeEntropy(stringProbs)
		}
	}

	return resultsByFile, nil
}
