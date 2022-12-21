package base64

import (
	"regexp"
)

var (
	// Adapted from https://stackoverflow.com/a/5885097 to only match base64 strings with at least 12 characters
	base64Regex = regexp.MustCompile("(?:[A-Za-z0-9+/]{4}){3,}(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=|[A-Za-z0-9+/]{4})")

	filterRegexes = []*regexp.Regexp{
		// non-hex letter
		regexp.MustCompile("[G-Zg-z]"),
		// uppercase letter
		regexp.MustCompile("[A-Z]"),
		// lowercase letter
		regexp.MustCompile("[a-z]"),
	}
)

/*
looksLikeActualBase64 uses some regex based heuristics to avoid false positive matching
of long words, hex strings, file paths, etc. as base64 strings, by checking that the
candidate string matches the regexes defined above
*/
func looksLikeActualBase64(candidate string) bool {
	for _, r := range filterRegexes {
		if !r.MatchString(candidate) {
			return false
		}
	}
	return true
}

/*
FindBase64Substrings returns a slice containing all the non-overlapping substrings of s
that are at least 12 characters long, and look like base64-encoded data. The function
uses regex-based heuristics to determine valid substrings but does not decode the data.
In particular, valid strings must have only valid base64 characters ([A-Za-z0-9+/]),
and have correct padding ('=') chars. Additionally, the following heuristics are used:

1. Valid strings contain one letter outside a-f (or A-F). This filters out hex literals.
2. Valid strings contain one uppercase letter and one lowercase letter

With a moderate sized input string, there will likely be some false positive matches.
Due to the 12 character minimum length, however, it is highly unlikely that a legitimate
base64 string will be excluded from the output. Additionally, if there are multiple
base64 encoded strings in the input, depending on how they are separated, they may
end up being concatenated together in a single element of the returned string slice.
*/
func FindBase64Substrings(s string) []string {
	matches := []string{}

	for _, candidate := range base64Regex.FindAllString(s, -1) {
		if looksLikeActualBase64(s) {
			matches = append(matches, candidate)
		}
	}
	return matches
}
