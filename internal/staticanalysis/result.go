package staticanalysis

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
)

/*
Result (staticanalysis.Result) is the top-level data structure that stores all data
produced by static analysis performed on a package / artifact. Each element
corresponds to an individual static analysis task (see Task). Note that this data
is sent across a sandbox boundary, so all nested structs must be JSON serialisable.
*/
type Result struct {
	// Note: the JSON names below match the values in task.go
	BasicData *BasicPackageData `json:"basic_data,omitempty"`

	ParsingData parsing.PackageResult `json:"parsing,omitempty"`

	ObfuscationData *obfuscation.Result `json:"obfuscation,omitempty"`
}

func (ar Result) String() string {
	parseDataStrings := make([]string, 0)
	for filename, parseData := range ar.ParsingData {
		parseDataStrings = append(parseDataStrings, fmt.Sprintf("== %s ==\n%s", filename, parseData))
	}

	parts := []string{
		fmt.Sprintf("File basic data\n%v", ar.BasicData),
		fmt.Sprintf("File parse data\n%s", strings.Join(parseDataStrings, "\n\n")),
		fmt.Sprintf("File obfuscation data\n%s", ar.ObfuscationData),
	}

	return strings.Join(parts, "\n\n########################\n\n")
}
