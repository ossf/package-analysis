package staticanalysis

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/basicdata"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/signals"
	"github.com/ossf/package-analysis/internal/utils"
)

/*
Result (staticanalysis.Result) is the top-level data structure that stores all data
produced by static analysis performed on a package / artifact. Each element
corresponds to an individual static analysis task (see Task).
*/
type Result struct {
	Files []SingleResult `json:"files"`
}

/*
SingleResult (staticanalysis.SingleResult) stores all data obtained by static analysis,
performed on a single file of a package / artifact. Each field corresponds to a different
analysis task (see Task). All nested structs must be JSON serialisable, so they can be
sent across the sandbox boundary.
*/
type SingleResult struct {
	// Filename is the relative path to the file within the package
	Filename string `json:"filename"`

	// NOTE: the JSON names below should match the values in task.go

	Basic   basicdata.FileData     `json:"basic,omitempty"`
	Parsing []parsing.SingleResult `json:"parsing,omitempty"`
	Signals signals.FileSignals    `json:"signals,omitempty"`
}

func (r SingleResult) String() string {
	parsingDataStrings := utils.Transform(r.Parsing, func(d parsing.SingleResult) string { return d.String() })

	parts := []string{
		fmt.Sprintf("==== SingleResult: %s ====\n", r.Filename),
		fmt.Sprintf("== basic data == \n%v", r.Basic),
		fmt.Sprintf("== parse data ==\n%s", strings.Join(parsingDataStrings, "\n")),
		fmt.Sprintf("== signals == \n%s", r.Signals),
	}

	return strings.Join(parts, "\n\n")
}
