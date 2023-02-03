package staticanalysis

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
)

// Result stores all data from static analysis performed on a package / artifact
// Note that this data is sent across a sandbox boundary, so must be serialisable.
type Result struct {
	// BasicData maps package file names to the BasicData collected for that file
	BasicData map[string]BasicFileData

	// ParseData maps package file names to the parsing result for that file, which
	// is parsing.Data for all attempted parsing languages. Currently, JavaScript is
	// the only supported parsing language.
	ParseData map[string]parsing.Result

	// Obfuscation holds all the data relating to obfuscation detection.
	Obfuscation *obfuscation.Result
}

func (ar Result) String() string {
	fileDataStrings := make([]string, 0)
	fileSignalsStrings := make([]string, 0)

	for filename, rawData := range ar.ParseData {
		fileDataStrings = append(fileDataStrings, fmt.Sprintf("== %s ==\n%s", filename, rawData))
	}
	for filename, signals := range ar.Obfuscation.Signals {
		fileSignalsStrings = append(fileSignalsStrings, fmt.Sprintf("== %s ==\n%s\n", filename, signals))
	}

	parts := []string{
		fmt.Sprintf("File basic data\n%v", ar.BasicData),
		fmt.Sprintf("File parse data\n%s", strings.Join(fileDataStrings, "\n\n")),
		fmt.Sprintf("File obfuscation data\n%s", strings.Join(fileSignalsStrings, "\n\n")),
	}

	return strings.Join(parts, "\n\n########################\n\n")
}
