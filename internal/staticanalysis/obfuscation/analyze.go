package obfuscation

import "github.com/ossf/package-analysis/internal/staticanalysis/parsing"

/*
Analyze performs obfuscation analysis for a package, operating on the data
obtained from parsing each file in the package.
If Language is empty (== NoLanguage), it means that the file could not be
parsed with any parser.
*/
func Analyze(parseData []*parsing.SingleResult) Result {
	result := Result{
		ExcludedFiles: []string{},
		Signals:       []FileSignals{},
	}

	for _, fileData := range parseData {
		switch fileData.Language {
		case parsing.NoLanguage:
			// couldn't be parsed
			result.ExcludedFiles = append(result.ExcludedFiles, fileData.Filename)
		case parsing.JavaScript:
			signals := ComputeFileSignals(*fileData)
			signals.Filename = fileData.Filename
			result.Signals = append(result.Signals, signals)
		}
	}

	return result
}
