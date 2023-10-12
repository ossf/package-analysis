package staticanalysis

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/basicdata"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/signals"
	"github.com/ossf/package-analysis/pkg/api/analysisrun"
)

// Result (staticanalysis.Result) is the top-level internal data structure
// that stores all data produced by static analysis performed on a package artifact.
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
	Filename string

	Basic   *basicdata.FileData
	Parsing *parsing.SingleResult
	Signals *signals.FileSignals
}

func (r SingleResult) String() string {
	parts := []string{
		fmt.Sprintf("==== SingleResult: %s ====\n", r.Filename),
		fmt.Sprintf("== basic data == \n%v", r.Basic),
		fmt.Sprintf("== parse data ==\n%v", r.Parsing),
		fmt.Sprintf("== signals == \n%s", r.Signals),
	}

	return strings.Join(parts, "\n\n")
}

// ProduceSerialisableResult converts the data in this Result object to the public API struct form
// analysisrun.StaticAnalysisResults.
// TODO unit test
func (r *Result) ProduceSerialisableResult() *analysisrun.StaticAnalysisResults {
	results := &analysisrun.StaticAnalysisResults{}

	for _, f := range r.Files {
		fr := analysisrun.StaticAnalysisFileResult{
			Filename: f.Filename,
		}
		if f.Basic != nil {
			fr.DetectedType = f.Basic.DetectedType
			fr.Size = f.Basic.Size
			fr.Sha256 = f.Basic.SHA256
			fr.LineLengths = f.Basic.LineLengths
		}
		if f.Parsing != nil && f.Parsing.Language == parsing.JavaScript {
			fr.Js = analysisrun.StaticAnalysisJsData{
				Identifiers:    f.Parsing.Identifiers,
				StringLiterals: f.Parsing.StringLiterals,
				IntLiterals:    f.Parsing.IntLiterals,
				FloatLiterals:  f.Parsing.FloatLiterals,
				Comments:       f.Parsing.Comments,
			}
		}
		if f.Signals != nil {

		}
	}

	return results
}
