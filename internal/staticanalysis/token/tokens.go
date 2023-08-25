package token

type Identifier struct {
	Name    string
	Type    IdentifierType
	Entropy float64
}

type Comment struct {
	Value string
}

type String struct {
	Value   string
	Raw     string
	Entropy float64
}

type Int struct {
	Value int64
	Raw   string
}

type Float struct {
	Value float64
	Raw   string
}
