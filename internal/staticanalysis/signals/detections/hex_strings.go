package detections

import (
	"regexp"
)

var hexRegex = regexp.MustCompile("[[:xdigit:]]{8,}")

/*
FindHexSubstrings returns all non-overlapping substrings of s
made up of at least 8 consecutive hexadecimal digits.
The leading 0x is not counted.
*/
func FindHexSubstrings(s string) []string {
	return hexRegex.FindAllString(s, -1)
}
