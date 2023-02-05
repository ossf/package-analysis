package parsing

import (
	"fmt"

	"github.com/ossf/package-analysis/internal/staticanalysis/token"
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

// parserOutput holds intermediate data from language-specific parsing functions
type parserOutput struct {
	ValidInput  bool
	Identifiers []parsedIdentifier
	Literals    []parsedLiteral[any]
	Comments    []parsedComment
}

var InvalidInput = parserOutput{ValidInput: false}
