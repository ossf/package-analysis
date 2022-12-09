package parsing

import (
	"fmt"
)

type SymbolType string
type IdentifierType string

const (
	Identifier SymbolType = "Identifier" // source code identifier (variable, class, function name)
	Literal    SymbolType = "Literal"    // source code data (string, integer, floating point literals)
	Comment    SymbolType = "Comment"    // source code comments
	Info       SymbolType = "Info"       // information about the parsing (e.g. number of bytes read by parser)
	Error      SymbolType = "Error"      // any error encountered by parser; some are recoverable and some are not

	Function       IdentifierType = "Function"       // function declaration / definition
	Variable       IdentifierType = "Variable"       // variable declaration / definition
	Parameter      IdentifierType = "Parameter"      // parameters to functions, constructors, catch blocks
	Class          IdentifierType = "Class"          // class declaration / definition
	Member         IdentifierType = "Member"         // access/mutation of an object member
	Property       IdentifierType = "Property"       // declaration of class property
	StatementLabel IdentifierType = "StatementLabel" // loop label
	Other          IdentifierType = "Other"          // the parser picked up that isn't accounted for above
	Unknown        IdentifierType = "Unknown"
)

var allTypes = []IdentifierType{
	Function,
	Variable,
	Parameter,
	Member,
	Property,
	Class,
	StatementLabel,
	Other,
	Unknown,
}

func CheckIdentifierType(s string) IdentifierType {
	for _, typeName := range allTypes {
		if s == string(typeName) {
			return typeName
		}
	}
	return Unknown
}

type ParsedIdentifier struct {
	Type IdentifierType
	Name string
	Pos  TextPosition
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
	Pos      TextPosition
}

func (l ParsedLiteral[T]) String() string {
	s := fmt.Sprintf("%s %v (%s) pos %d:%d", l.Type, l.Value, l.RawValue, l.Pos.Row(), l.Pos.Col())
	if l.InArray {
		s += " [array]"
	}
	return s
}

type ParsedComment struct {
	Type string
	Data string
	Pos  TextPosition
}

type ParseResult struct {
	ValidInput  bool
	Identifiers []ParsedIdentifier
	Literals    []ParsedLiteral[any]
	Comments    []ParsedComment
}

var InvalidInput = ParseResult{ValidInput: false}
