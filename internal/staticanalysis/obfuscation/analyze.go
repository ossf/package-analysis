package obfuscation

import "github.com/ossf/package-analysis/internal/staticanalysis/parsing"

// Analyze performs obfuscation analysis for a package, operating on the data
// obtained from parsing each file in the package
func Analyze(parseData parsing.PackageResult) Result {
	result := Result{
		SuspiciousFiles: map[string]string{},
		ExcludedFiles:   []string{},
		Signals:         map[string]FileSignals{},
	}

	for pathInArchive, data := range parseData {
		if data == nil || data[parsing.JavaScript] == nil {
			// couldn't be parsed
			result.ExcludedFiles = append(result.ExcludedFiles, pathInArchive)
		} else {
			signals := ComputeFileSignals(*data[parsing.JavaScript])
			// remove NaNs for JSON
			RemoveNaNs(&signals)
			result.Signals[pathInArchive] = signals

			// XXX these rules are extremely rudimentary
			if len(signals.SuspiciousIdentifiers["hex"]) > 4 {
				result.SuspiciousFiles[pathInArchive] = "hex identifiers"
			}

			if len(signals.SuspiciousIdentifiers["numeric"]) > 4 {
				result.SuspiciousFiles[pathInArchive] = "numeric identifiers"
			}

			longEscapedStringCount := 0
			for _, s := range signals.EscapedStrings {
				if len(s.RawLiteral) > 16 {
					longEscapedStringCount++
				}
			}
			if longEscapedStringCount >= 2 {
				result.SuspiciousFiles[pathInArchive] = "escaped strings"
			}
		}
	}

	return result
}
