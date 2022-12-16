package token

type IdentifierType string

const (
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
