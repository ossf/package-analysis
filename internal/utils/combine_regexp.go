package utils

import (
	"regexp"
	"strings"
)

// CombineRegexp creates a single regexp by joining the argument regexps together
// using the | operator. Each regexp is put into a separate non-capturing group before
// being combined.
func CombineRegexp(regexps ...*regexp.Regexp) *regexp.Regexp {
	patterns := Transform(regexps, func(r *regexp.Regexp) string {
		// create a non-capturing group for each regexp
		return "(?:" + r.String() + ")"
	})
	return regexp.MustCompile(strings.Join(patterns, "|"))
}
