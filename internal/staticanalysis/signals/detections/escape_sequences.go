package detections

import (
	"regexp"

	"github.com/ossf/package-analysis/pkg/api/staticanalysis/token"
)

/*
Escape sequences are defined by the regexes below. While octal, hex and
short/16-bit unicode escape sequences are mostly consistent across languages,
32-bit unicode (code point) escape sequences are more variable.
v1 appears in JS, PHP, Ruby while v2 appears in Python, C, Rust, Go.
*/
var (
	octalEscape       = regexp.MustCompile(`\\[0-7]{1,3}`)        // e.g "\077", "\251"
	hexEscape         = regexp.MustCompile(`\\x[[:xdigit:]]{2}`)  // e.g. "\x2a", "\x3f"
	unicodeEscape     = regexp.MustCompile(`\\u[[:xdigit:]]{4}`)  // e.g. "\u00af", "\u83bd"
	codePointEscapeV1 = regexp.MustCompile(`\\u\{[[:xdigit:]]+}`) // e.g. "\u{1ECC2}", \u{001FFF}"
	codePointEscapeV2 = regexp.MustCompile(`\\U[[:xdigit:]]{8}`)  // e.g. "\U0001ECC2", "\U00001FFF"

	allEscapeSequences = []*regexp.Regexp{octalEscape, hexEscape, unicodeEscape, codePointEscapeV1, codePointEscapeV2}
)

/*
IsHighlyEscaped returns true if a string literal exceeds the given
threshold count or frequency (in range [0, 1]) of escape sequences.

Supported escape sequences include:

 1. Octal escape: "\251",
 2. Hex escape: "\x3f",
 3. Unicode 16-bit escape: "\u103a",
 4. Unicode 32-bit escape: "\U00100FFF" or "\u{0100FF}".
*/
func IsHighlyEscaped(s token.String, thresholdCount int, thresholdFrequency float64) bool {
	escapeCount := 0

	for _, escapeSequencePattern := range allEscapeSequences {
		escapeCount += len(escapeSequencePattern.FindAllStringIndex(s.Raw, -1))
	}

	length := float64(len([]rune(s.Value))) // convert to rune slice first to count codepoints, not bytes
	return escapeCount >= thresholdCount || float64(escapeCount)/length >= thresholdFrequency
}
