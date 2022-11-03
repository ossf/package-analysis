package js

import (
	"fmt"
)

type IdentifierType string

const (
	Function       IdentifierType = "Function"
	Variable       IdentifierType = "Variable"
	Parameter      IdentifierType = "Parameter"
	Class          IdentifierType = "Class"
	Member         IdentifierType = "Member"
	Property       IdentifierType = "Property"
	StatementLabel IdentifierType = "StatementLabel"
	Other          IdentifierType = "Other"
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

func checkIdentifierType(s string) IdentifierType {
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

type ParseResult struct {
	Identifiers []ParsedIdentifier
	Literals    []ParsedLiteral[any]
}
