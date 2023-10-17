package token

import (
	"encoding/json"

	"golang.org/x/exp/maps"
)

// IdentifierType enumerates the possible types of a source code identifier,
// encountered during static analysis.
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

// MarshalJSON serializes this IdentifierType using its string representation
func (t IdentifierType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.String())
}

// UnmarshalJSON deserializes an IdentifierType serialized using MarshalJSON.
// If the supplied JSON contains an unrecognised name, the deserialised value is
// Unknown, and no error is returned.
func (t *IdentifierType) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}

	*t = ParseIdentifierType(name)
	return nil
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
