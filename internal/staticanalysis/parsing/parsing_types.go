package parsing

import (
	"fmt"
	"strings"

	"github.com/ossf/package-analysis/internal/staticanalysis/token"
	"github.com/ossf/package-analysis/internal/utils"
)

// Language represents a programming language used for parsing
type Language string

// SymbolType denotes a type of information collected during parsing.
// It may be a source code token (see token package), or status about the parsing process (info or error)
type SymbolType string

const (
	JavaScript Language = "JavaScript"

	Identifier SymbolType = "Identifier" // source code identifier (variable, class, function name)
	Literal    SymbolType = "Literal"    // source code data (string, integer, floating point literals)
	Comment    SymbolType = "Comment"    // source code comments
	Info       SymbolType = "Info"       // information about the parsing (e.g. number of bytes read by parser)
	Error      SymbolType = "Error"      // any error encountered by parser; some are recoverable and some are not
)

var allLanguages = []Language{JavaScript}

func SupportedLanguages() []Language {
	return allLanguages[:]
}

type parsedIdentifier struct {
	Type token.IdentifierType
	Name string
	Pos  token.Position
}

func (i parsedIdentifier) String() string {
	return fmt.Sprintf("%s %s [pos %d:%d]", i.Type, i.Name, i.Pos.Row(), i.Pos.Col())
}

type parsedLiteral[T any] struct {
	Type     string
	GoType   string
	Value    T
	RawValue string
	InArray  bool
	Pos      token.Position
}

func (l parsedLiteral[T]) String() string {
	s := fmt.Sprintf("%s (%s) %v (raw: %s) pos %d:%d", l.Type, l.GoType, l.Value, l.RawValue, l.Pos.Row(), l.Pos.Col())
	if l.InArray {
		s += " [array]"
	}
	return s
}

type parsedComment struct {
	Type string
	Data string
	Pos  token.Position
}

func (c parsedComment) String() string {
	return fmt.Sprintf("%s %s pos %d:%d", c.Type, c.Data, c.Pos.Row(), c.Pos.Col())
}

type parserStatus struct {
	Type    string
	Name    string
	Message string
	Pos     token.Position
}

func (s parserStatus) String() string {
	return fmt.Sprintf("[%s] %s: %s pos %d:%d", s.Type, s.Name, s.Message, s.Pos.Row(), s.Pos.Col())
}

// parseOutput holds intermediate data from language-specific parsing functions
type parseOutput struct {
	ValidInput  bool
	Identifiers []parsedIdentifier
	Literals    []parsedLiteral[any]
	Comments    []parsedComment
	Info        []parserStatus
	Errors      []parserStatus
}

func (p parseOutput) String() string {
	identifiers := utils.Transform(p.Identifiers, func(pi parsedIdentifier) string { return pi.String() })
	literals := utils.Transform(p.Literals, func(pl parsedLiteral[any]) string { return pl.String() })
	comments := utils.Transform(p.Comments, func(c parsedComment) string { return c.String() })
	info := utils.Transform(p.Info, func(i parserStatus) string { return i.String() })
	errors := utils.Transform(p.Errors, func(e parserStatus) string { return e.String() })

	parts := []string{
		"== Identifiers ==",
		strings.Join(identifiers, "\n"),
		"== Literals ==",
		strings.Join(literals, "\n"),
		"== Comments ==",
		strings.Join(comments, "\n"),
		"== Info ==",
		strings.Join(info, "\n"),
		"== Errors ==",
		strings.Join(errors, "\n"),
	}

	return strings.Join(parts, "\n")
}
