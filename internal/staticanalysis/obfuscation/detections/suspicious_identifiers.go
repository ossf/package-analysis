package detections

import "regexp"

var (
	hexIdentifier        = regexp.MustCompile("_0x[[:xdigit:]]{3,}")
	numericIdentifier    = regexp.MustCompile("^[A-Za-z_]\\d{3,}")
	singleCharIdentifier = regexp.MustCompile("^[A-Za-z_]$")
)

/*
SuspiciousIdentifierPatterns is a list of regex patterns to match source code
identifiers that are carry a suspicion of being obfuscated, due to being not
very human-friendly. A few matching identifiers may not indicate obfuscation,
but if there is a large number of suspicious identifiers (especially of the
same type) then obfuscation is probable.
*/
var SuspiciousIdentifierPatterns = map[string]*regexp.Regexp{
	"hex":     hexIdentifier,
	"numeric": numericIdentifier,
	"single":  singleCharIdentifier,
}
