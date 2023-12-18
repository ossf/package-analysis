package detections

import (
	"regexp"
	"strings"
)

var (
	// RFC4648 standard base 64 chars, padding optional, min length 16.
	standardBase64 = regexp.MustCompile("[[:alnum:]+/]{16,}(?:={0,2})?")
	// RFC4648 url/file-safe base 64 chars, padding optional, min length 16.
	urlSafeBase64 = regexp.MustCompile("[[:alnum:]-_]{16,}(?:={0,2})?")
	// Combines RFC4648 standard ('+', '/') + file-safe ('-', '_') base 64 variants.
	base64Regex = regexp.MustCompile(standardBase64.String() + "|" + urlSafeBase64.String())

	filterRegexes = []*regexp.Regexp{
		regexp.MustCompile("[[:upper:]]"),
		regexp.MustCompile("[[:lower:]]"),
		regexp.MustCompile("[G-Zg-z]"), // non-hex letter
	}
)

/*
looksLikeActualBase64 checks a candidate base64 string (that matches base64Regex)
using some rule-based heuristics to reduce false positive matching of e.g.
long words, hex strings, file paths. Additionally, if the candidate string
uses padding, its length is checked to ensure it is a multiple of 4 as required
by the Base64 standard.
*/
func looksLikeActualBase64(candidate string) bool {
	if strings.ContainsRune(candidate, '=') && len(candidate)%4 != 0 {
		return false
	}

	for _, r := range filterRegexes {
		if !r.MatchString(candidate) {
			return false
		}
	}

	return true
}

/*
FindBase64Substrings returns a slice containing all the non-overlapping substrings of s
that are at least 20 characters long, and look like base64-encoded data. The function
uses regex-based heuristics to determine valid substrings but does not decode the data.
In particular, valid strings must have only valid base64 characters ([A-Za-z0-9+/] or
[A-Za-z0-9-_], depending on the variant, plus up to 2 padding '=' characters).
If padding characters are included, then the string length must be a multiple of 4.

The following heuristic rules are checked to reduce the number of false positives.

1. Must have at least one uppercase letter
2. Must have at least one lowercase letter
3. Must have at least one letter outside A-F (or a-f) [this filters out hex strings]
4. If padding characters are included, the string length must be a multiple of 4

While false positive matches will occur, due to the minimum length requirement
it is highly unlikely that a legitimate base64 string will be excluded from the output.

Note that, if there are multiple base64 encoded strings in the input, depending
on how they are separated, they may end up being concatenated together into a single
string in the returned string slice.
*/
func FindBase64Substrings(s string) []string {
	matches := []string{}

	for _, candidate := range base64Regex.FindAllString(s, -1) {
		if looksLikeActualBase64(candidate) {
			matches = append(matches, candidate)
		}
	}
	return matches
}
