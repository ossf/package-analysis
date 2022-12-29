package escaping

import (
	"regexp"

	"github.com/texttheater/golang-levenshtein/levenshtein"

	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

const (
	minEscapeCharCount     = 8
	minEscapeCharFrequency = 0.25
)

var (
	// detects hex escape sequences ("\x3f") in raw strings
	hexEscapeRegex = regexp.MustCompile(`\\x[[:xdigit:]]{2}`)
	// detects unicode escape sequences ("\u00af") in raw strings
	unicodeEscapeRegex = regexp.MustCompile(`\\u[[:xdigit:]]{4}`)
)

/*
IsHighlyEscaped returns true if a string literal contains a high amount of escape characters,
which is defined by either of the following conditions being true:

1. At least 8 characters in the string are escaped
2. At least 25% of characters in the string are escaped

where escaped means either hex escaped ("\xfc") or unicode escaped ("\u00af") using escape literals
*/
func IsHighlyEscaped(s token.String) bool {
	hexEscapeChars := hexEscapeRegex.FindAllStringIndex(s.Raw, -1)
	unicodeEscapeChars := unicodeEscapeRegex.FindAllStringIndex(s.Raw, -1)
	totalEscapeChars := len(hexEscapeChars) + len(unicodeEscapeChars)
	length := float64(len([]rune(s.Value))) // convert to rune slice first to count codepoints, not bytes
	return totalEscapeChars >= minEscapeCharCount || float64(totalEscapeChars)/length >= minEscapeCharFrequency
}

/*
LevenshteinRatio returns the Levenshtein ratio between the parsed and raw versions
of a string literal. This quantity is defined for two strings 'source' and 'target' as

(sourceLength + targetLength - distance) / (sourceLength + targetLength)

where 'distance' refers to Levenshtein (edit) distance. The particular version of
Levenshtein distance used counts substitution as 2 operations (deletion and addition).
*/
func LevenshteinRatio(s token.String) float64 {
	source := []rune(s.Raw)
	target := []rune(s.Value)
	return levenshtein.RatioForStrings(source, target, levenshtein.DefaultOptions)
}
