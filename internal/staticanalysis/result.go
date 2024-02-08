package staticanalysis

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/basicdata"
	"github.com/ossf/package-analysis/internal/staticanalysis/parsing"
	"github.com/ossf/package-analysis/internal/staticanalysis/signals"
	"github.com/ossf/package-analysis/pkg/api/staticanalysis"
)

// Result (staticanalysis.Result) is the top-level internal data structure
// that stores all data produced by static analysis performed on a package artifact.
type Result struct {
	Archive ArchiveResult
	Files   []SingleResult
}

type ArchiveResult struct {
	// DetectedType records the output of the `file` command run on the archive.
	DetectedType string

	// Size records the (compressed) size of the archive (as reported by the filesystem).
	Size int64

	// SHA256 records the SHA256 hashsum of the archive.
	SHA256 string
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

// ToAPIResults converts the data in this Result object into the
// public staticanalysis.Results format defined in pkg/api/staticanalysis.
func (r *Result) ToAPIResults() *staticanalysis.Results {
	results := &staticanalysis.Results{}

	for _, f := range r.Files {
		fr := staticanalysis.FileResult{
			Filename: f.Filename,
		}
		if f.Basic != nil {
			fr.DetectedType = f.Basic.DetectedType
			fr.Size = f.Basic.Size
			fr.SHA256 = f.Basic.SHA256
			// only populate value counts if nonempty
			if f.Basic.LineLengths.Len() > 0 {
				fr.LineLengths = &f.Basic.LineLengths
			}
		}
		if f.Parsing != nil && f.Parsing.Language == parsing.JavaScript {
			fr.Js = &staticanalysis.JsData{
				Identifiers:    f.Parsing.Identifiers,
				StringLiterals: f.Parsing.StringLiterals,
				IntLiterals:    f.Parsing.IntLiterals,
				FloatLiterals:  f.Parsing.FloatLiterals,
				Comments:       f.Parsing.Comments,
			}
		}
		if f.Signals != nil {
			// only populate value counts if nonempty
			if f.Signals.IdentifierLengths.Len() > 0 {
				fr.IdentifierLengths = &f.Signals.IdentifierLengths
			}
			if f.Signals.StringLengths.Len() > 0 {
				fr.StringLengths = &f.Signals.StringLengths
			}
			fr.Base64Strings = f.Signals.Base64Strings
			fr.HexStrings = f.Signals.HexStrings
			fr.IPAddresses = f.Signals.IPAddresses
			fr.URLs = f.Signals.URLs
			fr.EscapedStrings = f.Signals.EscapedStrings
			fr.SuspiciousIdentifiers = f.Signals.SuspiciousIdentifiers
		}

		results.Files = append(results.Files, fr)
	}

	return results
}
