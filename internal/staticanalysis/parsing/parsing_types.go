package parsing

import (
	"fmt"

	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

// Language represents a programming language used for parsing
type Language string

// SymbolType denotes a type of information collected during parsing. It may be a source code token (see token package), or
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

type ParsedIdentifier struct {
	Type token.IdentifierType
	Name string
	Pos  token.Position
}

func (i ParsedIdentifier) String() string {
	return fmt.Sprintf("%s %s [pos %d:%d]", i.Type, i.Name, i.Pos.Row(), i.Pos.Col())
}

type ParsedLiteral[T any] struct {
	Type     string
	GoType   string
	Value    T
	RawValue string
	InArray  bool
	Pos      token.Position
}

func (l ParsedLiteral[T]) String() string {
	s := fmt.Sprintf("%s (%s) %v (raw: %s) pos %d:%d", l.Type, l.GoType, l.Value, l.RawValue, l.Pos.Row(), l.Pos.Col())
	if l.InArray {
		s += " [array]"
	}
	return s
}

type ParsedComment struct {
	Type string
	Data string
	Pos  token.Position
}

// internalResult holds intermediate data between parsing functions
type internalResult struct {
	ValidInput  bool
	Identifiers []ParsedIdentifier
	Literals    []ParsedLiteral[any]
	Comments    []ParsedComment
}

var InvalidInput = internalResult{ValidInput: false}
