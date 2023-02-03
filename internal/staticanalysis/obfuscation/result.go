package obfuscation

import "github.com/ossf/package-analysis/internal/staticanalysis/parsing"

type Result struct {
	// SuspiciousFiles lists all files in the package that are suspected to contain
	// obfuscated code. This should be treated not as an absolute determination but
	// more of a flag for human review.
	SuspiciousFiles []string

	// ExcludedFiles is a list of package files that were excluded from analysis,
	// e.g. because they could not be parsed by any supported parser
	ExcludedFiles []string

	// Signals maps package file names to corresponding obfuscation.FileSignals
	// that are used to generate the list of suspicious files
	Signals map[string]FileSignals
}

func ComputeResult(fileParseData map[string]parsing.Result) *Result {
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
