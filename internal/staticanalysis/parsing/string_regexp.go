package parsing

import (
	"os"
	"regexp"
	"strings"

	"github.com/ossf/package-analysis/internal/utils"
)

// General reference for matching string literals
// https://blog.stevenlevithan.com/archives/match-quoted-string

// https://stackoverflow.com/a/10786066
var singleQuotedString = regexp.MustCompile(`'[^'\\]*(\\.[^'\\]*)*'`)
var doubleQuotedString = regexp.MustCompile(`"[^"\\]*(\\.[^"\\]*)*"`)
var backTickQuotedString = regexp.MustCompile("`[^`\\\\]*(\\\\.[^`\\\\]*)*`")

// https://stackoverflow.com/a/30737232
var singleQuotedString2 = regexp.MustCompile(`'(?:[^'\\]*(?:\\.)?)*'`)
var doubleQuotedString2 = regexp.MustCompile(`"(?:[^"\\]*(?:\\.)?)*"`)
var backTickQuotedString2 = regexp.MustCompile("`(?:[^`\\\\]*(?:\\\\.)?)*`")

func combineRegexp(regexps ...*regexp.Regexp) *regexp.Regexp {
	patterns := utils.Transform(regexps, func(r *regexp.Regexp) string { return r.String() })
	return regexp.MustCompile(strings.Join(patterns, "|"))
}

//goland:noinspection GoUnusedGlobalVariable
var anyQuotedString = combineRegexp(singleQuotedString, doubleQuotedString, backTickQuotedString)

//goland:noinspection GoUnusedGlobalVariable
var anyQuotedString2 = combineRegexp(singleQuotedString2, doubleQuotedString2, backTickQuotedString2)

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

	unquotedStrings := utils.Transform(allStrings, dequote)

	if allStrings != nil {
		return &ExtractedStrings{Strings: unquotedStrings, RawLiterals: allStrings}, nil
	} else {
		return &ExtractedStrings{Strings: []string{}, RawLiterals: []string{}}, nil
	}
}

func FindStringsInFile(filePath string, stringRegexp *regexp.Regexp) (*ExtractedStrings, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	fileString := string(fileBytes)
	return FindStringsInCode(fileString, stringRegexp)
}
