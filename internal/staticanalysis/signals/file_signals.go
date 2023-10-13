package signals

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/pkg/api/staticanalysis"
	"github.com/ossf/package-analysis/pkg/valuecounts"
)

// FileSignals holds information related to the presence of obfuscated code in a single file.
type FileSignals struct {
	// The following two variables respectively record how many string literals
	// and identifiers in the file have a given length. The absence of a count
	// for a particular lengths means that there were no symbols of that length
	// in the file.
	IdentifierLengths valuecounts.ValueCounts
	StringLengths     valuecounts.ValueCounts

	// SuspiciousIdentifiers holds identifiers that are deemed 'suspicious' (i.e.
	// indicative of obfuscation) according to certain rules. Each entry contains
	// the identifier name and the name of the first rule it was matched against.
	SuspiciousIdentifiers []staticanalysis.SuspiciousIdentifier

	// EscapedStrings contain string literals that contain large amount of escape
	// characters, which may indicate obfuscation.
	EscapedStrings []staticanalysis.EscapedString

	// Base64Strings holds a list of (substrings of) string literals found in the
	// file that match a base64 regex pattern. This patten has a minimum matching
	// length in order to reduce the number of false positives.
	Base64Strings []string

	// HexStrings holds a list of (substrings of) string literals found in the
	// file that contain long (>8 digits) hexadecimal digit sequences.
	HexStrings []string

	// IPAddresses contains any IP addresses found in string literals
	IPAddresses []string

	// URLs contains any urls (http or https) found in string literals
	URLs []string
}

func (s FileSignals) String() string {
	parts := []string{
		fmt.Sprintf("identifier length counts: %v", s.IdentifierLengths),
		fmt.Sprintf("string length counts: %v", s.StringLengths),

		fmt.Sprintf("suspicious identifiers: %v", s.SuspiciousIdentifiers),
		fmt.Sprintf("escaped strings: %v", s.EscapedStrings),
		fmt.Sprintf("potential base64 strings: %v", s.Base64Strings),
		fmt.Sprintf("hex strings: %v", s.HexStrings),
		fmt.Sprintf("IP addresses: %v", s.IPAddresses),
		fmt.Sprintf("URLs: %v", s.URLs),
	}
	return strings.Join(parts, "\n")
}
