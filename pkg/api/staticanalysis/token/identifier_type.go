package token

import (
	"golang.org/x/exp/maps"
)

type IdentifierType int

const (
	Unknown        IdentifierType = iota
	Function                      // function declaration / definition
	Variable                      // variable declaration / definition
	Parameter                     // parameters to functions, constructors, catch blocks
	Class                         // class declaration / definition
	Member                        // access/mutation of an object member
	Property                      // declaration of class property
	StatementLabel                // loop label
	Other                         // something the parser picked up that isn't accounted for above
)

var stringValues = map[IdentifierType]string{
	Unknown:        "Unknown",
	Function:       "Function",
	Variable:       "Variable",
	Parameter:      "Parameter",
	Class:          "Class",
	Member:         "Member",
	Property:       "Property",
	StatementLabel: "StatementLabel",
	Other:          "Other",
}

func (t IdentifierType) String() string {
	return stringValues[t]
}

func IdentifierTypes() []IdentifierType {
	return maps.Keys(stringValues)
}

func ParseIdentifierType(s string) IdentifierType {
	for name, stringVal := range stringValues {
		if s == stringVal {
			return name
		}
	}
	return Unknown
}
