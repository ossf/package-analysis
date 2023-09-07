package signals

import "github.com/ossf/package-analysis/internal/staticanalysis/parsing"

/*
Analyze performs signals analysis for a package, operating on the data
obtained from parsing each file in the package.
If Language is empty (== NoLanguage), it means that the file could not be
parsed with any parser.
*/
func Analyze(parseData []parsing.SingleResult) Result {
	result := Result{
		Files: []FileSignals{},
	}

	for _, d := range parseData {
		signalsData := ComputeFileSignals(d)
		signalsData.Filename = d.Filename
		result.Files = append(result.Files, signalsData)
	}

	return result
}
