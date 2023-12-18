package signals

import (
	"unicode/utf8"

	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/signals/detections"
	"github.com/ossf/package-analysis/internal/utils"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis/token"
	"github.com/ossf/package-analysis/pkg/valuecounts"
)

// countLengths returns a map containing the aggregated lengths
// of each of the strings in the input list
func countLengths(symbols []string) valuecounts.ValueCounts {
	lengths := make([]int, 0, len(symbols))
	for _, s := range symbols {
		lengths = append(lengths, utf8.RuneCountInString(s))
	}

	return valuecounts.Count(lengths)
}

// AnalyzeSingle collects signals of interest for a file in a package, operating on a single
// parsing result (i.e. from one language parser). It returns a FileSignals object, containing
// information that may be useful to determine whether the file contains malicious code.
func AnalyzeSingle(parseData parsing.SingleResult) FileSignals {
	identifierNames := utils.Transform(parseData.Identifiers, func(i token.Identifier) string { return i.Name })
	stringLiterals := utils.Transform(parseData.StringLiterals, func(s token.String) string { return s.Value })

	identifierLengths := countLengths(identifierNames)
	stringLengths := countLengths(stringLiterals)

	signals := FileSignals{
		IdentifierLengths:     identifierLengths,
		StringLengths:         stringLengths,
		Base64Strings:         []string{},
		HexStrings:            []string{},
		EscapedStrings:        []staticanalysis.EscapedString{},
		SuspiciousIdentifiers: []staticanalysis.SuspiciousIdentifier{},
		URLs:                  []string{},
		IPAddresses:           []string{},
	}

	for _, name := range identifierNames {
		for rule, pattern := range detections.SuspiciousIdentifierPatterns {
			if pattern.MatchString(name) {
				signals.SuspiciousIdentifiers = append(signals.SuspiciousIdentifiers, staticanalysis.SuspiciousIdentifier{name, rule})
				break // don't bother searching for multiple matching rules
			}
		}
	}

	for _, sl := range parseData.StringLiterals {
		signals.Base64Strings = append(signals.Base64Strings, detections.FindBase64Substrings(sl.Value)...)
		signals.HexStrings = append(signals.HexStrings, detections.FindHexSubstrings(sl.Value)...)
		signals.URLs = append(signals.URLs, detections.FindURLs(sl.Value)...)
		signals.IPAddresses = append(signals.IPAddresses, detections.FindIPAddresses(sl.Value)...)
		if detections.IsHighlyEscaped(sl, 8, 0.25) {
			escapedString := staticanalysis.EscapedString{
				Value:           sl.Value,
				Raw:             sl.Raw,
				LevenshteinDist: sl.LevenshteinDist(),
			}
			signals.EscapedStrings = append(signals.EscapedStrings, escapedString)
		}
	}

	return signals
}
