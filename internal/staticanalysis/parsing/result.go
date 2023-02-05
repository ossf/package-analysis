package parsing

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

// PackageResult holds parsing data for all files in a package, mapping
// file paths to the FileResult for that file
type PackageResult map[string]FileResult

// FileResult holds the full set of parsing data for a single file, which may
// contain valid parsing results for multiple languages
type FileResult map[Language]*SingleResult

// SingleResult holds information about source code tokens found in a file,
// when parsed as particular language.
type SingleResult struct {
	Identifiers    []token.Identifier `json:"identifiers"`
	StringLiterals []token.String     `json:"string_literals"`
	IntLiterals    []token.Int        `json:"int_literals"`
	FloatLiterals  []token.Float      `json:"float_literals"`
	Comments       []token.Comment    `json:"comments"`
}

func (rd SingleResult) String() string {
	parts := []string{
		fmt.Sprintf("identifiers\n%v\n", rd.Identifiers),
		fmt.Sprintf("string literals\n%v\n", rd.StringLiterals),
		fmt.Sprintf("integer literals\n%v\n", rd.IntLiterals),
		fmt.Sprintf("float literals\n%v\n", rd.FloatLiterals),
		fmt.Sprintf("comments\n%v\n", rd.Comments),
	}
	return strings.Join(parts, "\n-------------------\n")
}
