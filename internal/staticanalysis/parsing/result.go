package parsing

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/pkg/api/staticanalysis/token"
)

// SingleResult holds processed information about source code tokens
// found in a single file by a single language parser
type SingleResult struct {
	Language       Language           `json:"language"`
	Identifiers    []token.Identifier `json:"identifiers"`
	StringLiterals []token.String     `json:"string_literals"`
	IntLiterals    []token.Int        `json:"int_literals"`
	FloatLiterals  []token.Float      `json:"float_literals"`
	Comments       []token.Comment    `json:"comments"`
	// future: external function calls / references (e.g. eval)
}

func (r SingleResult) String() string {
	parts := []string{
		fmt.Sprintf("language: %s", r.Language),
		fmt.Sprintf("identifiers\n%v", r.Identifiers),
		fmt.Sprintf("string literals\n%v", r.StringLiterals),
		fmt.Sprintf("integer literals\n%v", r.IntLiterals),
		fmt.Sprintf("float literals\n%v", r.FloatLiterals),
		fmt.Sprintf("comments\n%v", r.Comments),
	}
	return strings.Join(parts, "\n")
}
