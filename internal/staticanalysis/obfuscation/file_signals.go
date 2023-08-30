package obfuscation

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/detections"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/token"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/internal/utils/valuecounts"
)

// FileSignals holds information related to the presence of obfuscated code in a single file.
type FileSignals struct {
	// Filename is the path in the package
	Filename string `json:"filename"`

	// The following two variables respectively record how many string literals
	// and identifiers in the file have a given length. The absence of a count
	// for a particular lengths means that there were no symbols of that length
	// in the file.
	IdentifierLengths valuecounts.ValueCounts `json:"identifier_lengths"`
	StringLengths     valuecounts.ValueCounts `json:"string_lengths"`

	// SuspiciousIdentifiers holds identifiers that are deemed 'suspicious' (i.e.
	// indicative of obfuscation) according to certain rules. Each entry contains
	// the identifier name and the name of the first rule it was matched against.
	SuspiciousIdentifiers []SuspiciousIdentifier `json:"suspicious_identifiers"`

	// EscapedStrings contain string literals that contain large amount of escape
	// characters, which may indicate obfuscation.
	EscapedStrings []EscapedString `json:"escaped_strings"`

	// Base64Strings holds a list of (substrings of) string literals found in the
	// file that match a base64 regex pattern. This patten has a minimum matching
	// length in order to reduce the number of false positives.
	Base64Strings []string `json:"base64_strings"`

	// EmailAddresses contains any email addresses found in string literals
	EmailAddresses []string `json:"email_addresses"`

	// HexStrings holds a list of (substrings of) string literals found in the
	// file that contain long (>8 digits) hexadecimal digit sequences.
	HexStrings []string `json:"hex_strings"`

	// IPAddresses contains any IP addresses found in string literals
	IPAddresses []string `json:"ip_addresses"`

	// URLs contains any urls (http or https) found in string literals
	URLs []string `json:"urls"`
}

func (s FileSignals) String() string {
	parts := []string{
		fmt.Sprintf("filename: %s", s.Filename),
		fmt.Sprintf("identifier length counts: %v", s.IdentifierLengths),
		fmt.Sprintf("string length counts: %v", s.StringLengths),

		fmt.Sprintf("suspicious identifiers: %v", s.SuspiciousIdentifiers),
		fmt.Sprintf("escaped strings: %v", s.EscapedStrings),
		fmt.Sprintf("potential base64 strings: %v", s.Base64Strings),
		fmt.Sprintf("hex strings: %v", s.HexStrings),
		fmt.Sprintf("email addresses: %v", s.EmailAddresses),
		fmt.Sprintf("hex strings: %v", s.HexStrings),
		fmt.Sprintf("IP addresses: %v", s.IPAddresses),
		fmt.Sprintf("URLs: %v", s.URLs),
	}
	return strings.Join(parts, "\n")
}

type EscapedString struct {
	RawLiteral       string  `json:"raw_literal"`
	LevenshteinRatio float64 `json:"levenshtein_ratio"`
}

// SuspiciousIdentifier is an identifier that matches a specific rule intended
// to pick out (potentially) suspicious names. Name stores the actual identifier,
// and Rule holds the rule that the identifier matched against.
type SuspiciousIdentifier struct {
	Name string `json:"name"`
	Rule string `json:"rule"`
}

// countLengths returns a map containing the aggregated lengths
// of each of the strings in the input list
func countLengths(symbols []string) valuecounts.ValueCounts {
	lengths := make([]int, 0, len(symbols))
	for _, s := range symbols {
		lengths = append(lengths, utf8.RuneCountInString(s))
	}

	return valuecounts.Count(lengths)
}

// ComputeFileSignals creates a FileSignals object based on the parsing data obtained from
// a given file. These signals may be useful to determine whether the code is obfuscated.
func ComputeFileSignals(parseData parsing.SingleResult) FileSignals {
	identifierNames := utils.Transform(parseData.Identifiers, func(i token.Identifier) string { return i.Name })
	stringLiterals := utils.Transform(parseData.StringLiterals, func(s token.String) string { return s.Value })

	identifierLengths := countLengths(identifierNames)
	stringLengths := countLengths(stringLiterals)

	signals := FileSignals{
		IdentifierLengths:     identifierLengths,
		StringLengths:         stringLengths,
		Base64Strings:         []string{},
		HexStrings:            []string{},
		EscapedStrings:        []EscapedString{},
		SuspiciousIdentifiers: []SuspiciousIdentifier{},
		URLs:                  []string{},
		IPAddresses:           []string{},
		EmailAddresses:        []string{},
	}

	for _, name := range identifierNames {
		for rule, pattern := range detections.SuspiciousIdentifierPatterns {
			if pattern.MatchString(name) {
				signals.SuspiciousIdentifiers = append(signals.SuspiciousIdentifiers, SuspiciousIdentifier{name, rule})
				break // don't bother searching for multiple matching rules
			}
		}
	}

	for _, sl := range parseData.StringLiterals {
		signals.Base64Strings = append(signals.Base64Strings, detections.FindBase64Substrings(sl.Value)...)
		signals.HexStrings = append(signals.HexStrings, detections.FindHexSubstrings(sl.Value)...)
		signals.URLs = append(signals.URLs, detections.FindURLs(sl.Value)...)
		signals.IPAddresses = append(signals.IPAddresses, detections.FindIPAddresses(sl.Value)...)
		signals.EmailAddresses = append(signals.EmailAddresses, detections.FindEmailAddresses(sl.Value)...)
		if detections.IsHighlyEscaped(sl, 8, 0.25) {
			escapedString := EscapedString{
				RawLiteral:       sl.Raw,
				LevenshteinRatio: detections.LevenshteinRatio(sl),
			}
			signals.EscapedStrings = append(signals.EscapedStrings, escapedString)
		}
	}

	return signals
}
