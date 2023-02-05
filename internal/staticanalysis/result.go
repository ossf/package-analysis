package staticanalysis

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
)

/*
Result (staticanalysis.Result) is the top-level data structure that stores all data
produced by static analysis performed on a package / artifact. Each element
corresponds to an individual static analysis task (see Task). Note that this data
is sent across a sandbox boundary, so all nested structs must be JSON serialisable.
*/
type Result map[Task]any

func (ar Result) String() string {
	var parts []string

	if ar[BasicData] != nil {
		parts = append(parts, fmt.Sprintf("File basic data\n%v", ar[BasicData]))
	}

	if ar[Parsing] != nil {
		parseDataStrings := make([]string, 0)
		if parseData, ok := ar[Parsing].(parsing.PackageResult); ok {
			for filename, rawData := range parseData {
				parseDataStrings = append(parseDataStrings, fmt.Sprintf("== %s ==\n%s", filename, rawData))
			}
			parts = append(parts, fmt.Sprintf("File parse data\n%s", strings.Join(parseDataStrings, "\n\n")))
		}
	}

	if ar[Obfuscation] != nil {
		parts = append(parts, fmt.Sprintf("File obfuscation data\n%s", ar[Obfuscation]))
	}

	return strings.Join(parts, "\n\n########################\n\n")
}
