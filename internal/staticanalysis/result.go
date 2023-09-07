package staticanalysis

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/obfuscation"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/utils"
)

/*
Result (staticanalysis.Result) is the top-level data structure that stores all data
produced by static analysis performed on a package / artifact. Each element
corresponds to an individual static analysis task (see Task). Note that this data
is sent across a sandbox boundary, so all nested structs must be JSON serialisable.
*/
type Result struct {
	// NOTE: the JSON names below should match the values in task.go
	BasicData *BasicPackageData `json:"basic,omitempty"`

	ParsingData []parsing.SingleResult `json:"parsing,omitempty"`

	ObfuscationData *obfuscation.Result `json:"obfuscation,omitempty"`
}

func (ar Result) String() string {
	parsingDataStrings := utils.Transform(ar.ParsingData, func(d parsing.SingleResult) string { return d.String() })

	parts := []string{
		fmt.Sprintf("File basic data\n%v", ar.BasicData),
		fmt.Sprintf("File parse data\n%s", strings.Join(parsingDataStrings, "\n\n")),
		fmt.Sprintf("File obfuscation data\n%s", ar.ObfuscationData),
	}

	return strings.Join(parts, "\n\n########################\n\n")
}
