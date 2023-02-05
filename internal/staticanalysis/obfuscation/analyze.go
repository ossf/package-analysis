package obfuscation

import "github.com/ossf/package-analysis/internal/staticanalysis/parsing"

// Analyze performs obfuscation analysis for a package, operating on the data
// obtained from parsing each file in the package
func Analyze(fileParseData parsing.PackageResult) *Result {
	result := &Result{
		SuspiciousFiles: []string{},
		ExcludedFiles:   []string{},
		Signals:         map[string]FileSignals{},
	}

	for pathInArchive, parseData := range fileParseData {
		if parseData == nil || parseData[parsing.JavaScript] == nil {
			// couldn't be parsed
			result.ExcludedFiles = append(result.ExcludedFiles, pathInArchive)
		} else {
			signals := ComputeFileSignals(*parseData[parsing.JavaScript])
			// remove NaNs for JSON
			RemoveNaNs(&signals)
			result.Signals[pathInArchive] = signals
		}
	}

	return result
}
