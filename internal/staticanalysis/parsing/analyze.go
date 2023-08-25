package parsing

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ossf/package-analysis/internal/staticanalysis/externalcmd"
	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stringentropy"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

func processJsData(filename string, fileData singleParseData) *SingleResult {
	if !fileData.ValidInput {
		return &SingleResult{
			Filename: filename,
			Language: NoLanguage,
		}
	}

	result := &SingleResult{
		Filename: filename,
		Language: JavaScript,
		// Initialise with empty slices to avoid null values in JSON
		Identifiers:    []token.Identifier{},
		StringLiterals: []token.String{},
		IntLiterals:    []token.Int{},
		FloatLiterals:  []token.Float{},
		Comments:       []token.Comment{},
	}

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
		case token.Function, token.Class, token.Parameter, token.Property, token.Variable:
			result.Identifiers = append(result.Identifiers, token.Identifier{Name: ident.Name, Type: ident.Type})
		}
	}

	for _, c := range fileData.Comments {
		result.Comments = append(result.Comments, token.Comment{Value: c.Data})
	}
	return result
}

// computeEntropy populates entropy values for identifiers and string literals.
// Character frequency distribution is estimated by aggregating character counts
// in across all identifiers and string literals respectively in the package.
func computeEntropy(parseResults []*SingleResult) {
	var strings []string
	var identifiers []string

	for _, result := range parseResults {
		for _, sl := range result.StringLiterals {
			strings = append(strings, sl.Value)
		}
		for _, id := range result.Identifiers {
			identifiers = append(identifiers, id.Name)
		}
	}

	stringLiteralCharDistribution := stringentropy.CharacterProbabilities(strings)
	identifierCharDistribution := stringentropy.CharacterProbabilities(identifiers)

	for _, result := range parseResults {
		for _, sl := range result.StringLiterals {
			sl.Entropy = stringentropy.CalculateEntropy(sl.Value, stringLiteralCharDistribution)
		}
		for _, id := range result.Identifiers {
			id.Entropy = stringentropy.CalculateEntropy(id.Name, identifierCharDistribution)
		}
	}
}

/*
Analyze (parsing.Analyze) parses the specified list of files using all supported parsers
and returns a list of parsing.SingleResult, each of which holds information about source code
tokens found for a given file and langauge parser combination.

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
func Analyze(parserConfig ParserConfig, input externalcmd.Input, printDebug bool) ([]*SingleResult, error) {
	// JavaScript parsing
	jsResults, rawOutput, err := parseJS(parserConfig, input)
	if printDebug {
		fmt.Fprintf(os.Stderr, "\nRaw JSON:\n%s\n", rawOutput)
	}
	if err != nil {
		return nil, err
	}

	packageResult := make([]*SingleResult, 0, len(jsResults))
	for filename, jsData := range jsResults {
		packageResult = append(packageResult, processJsData(filename, jsData))
	}

	computeEntropy(packageResult)

	return packageResult, nil
}
