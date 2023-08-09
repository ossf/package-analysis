package parsing

import (
	"os"
	"regexp"

	"github.com/ossf/package-analysis/internal/utils"
)

// General reference for matching string literals
// https://blog.stevenlevithan.com/archives/match-quoted-string

// https://stackoverflow.com/a/10786066
var (
	singleQuotedString   = regexp.MustCompile(`'[^'\\]*(\\.[^'\\]*)*'`)
	doubleQuotedString   = regexp.MustCompile(`"[^"\\]*(\\.[^"\\]*)*"`)
	backTickQuotedString = regexp.MustCompile("`[^`\\\\]*(\\\\.[^`\\\\]*)*`")
)

// https://stackoverflow.com/a/30737232
var (
	singleQuotedString2   = regexp.MustCompile(`'(?:[^'\\]*(?:\\.)?)*'`)
	doubleQuotedString2   = regexp.MustCompile(`"(?:[^"\\]*(?:\\.)?)*"`)
	backTickQuotedString2 = regexp.MustCompile("`(?:[^`\\\\]*(?:\\\\.)?)*`")
)

//goland:noinspection GoUnusedGlobalVariable
var anyQuotedString = utils.CombineRegexp(singleQuotedString, doubleQuotedString, backTickQuotedString)

//goland:noinspection GoUnusedGlobalVariable
var anyQuotedString2 = utils.CombineRegexp(singleQuotedString2, doubleQuotedString2, backTickQuotedString2)

type ExtractedStrings struct {
	RawLiterals []string
	Strings     []string
}

func dequote(s string) string {
	if len(s) <= 2 {
		return ""
	} else {
		return s[1 : len(s)-1]
	}
}

func FindStringsInCode(source string, stringRegexp *regexp.Regexp) (*ExtractedStrings, error) {
	allStrings := stringRegexp.FindAllString(source, -1)
	if allStrings == nil {
		return &ExtractedStrings{Strings: []string{}, RawLiterals: []string{}}, nil
	}

	unquotedStrings := utils.Transform(allStrings, dequote)
	return &ExtractedStrings{Strings: unquotedStrings, RawLiterals: allStrings}, nil
}

func FindStringsInFile(filePath string, stringRegexp *regexp.Regexp) (*ExtractedStrings, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileString := string(fileBytes)
	return FindStringsInCode(fileString, stringRegexp)
}
