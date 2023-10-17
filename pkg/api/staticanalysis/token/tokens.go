package token

import (
	"github.com/texttheater/golang-levenshtein/levenshtein"

	"github.com/ossf/package-analysis/internal/staticanalysis/signals/stringentropy"
)

// Identifier records some kind of user-defined symbol name in source code.
// Valid types of identifier are defined using IdentifierType.
type Identifier struct {
	Name    string         `json:"name"`
	Type    IdentifierType `json:"type"`
	Entropy float64        `json:"entropy"`
}

// ComputeEntropy computes the entropy of this identifier's name under the given
// character distribution, and sets its Entropy field to the resulting value.
func (i *Identifier) ComputeEntropy(probs map[rune]float64) {
	i.Entropy = stringentropy.Calculate(i.Name, probs)
}

// String records a string literal occurring in the source code.
type String struct {
	Value   string  `json:"value"`
	Raw     string  `json:"raw"`
	Entropy float64 `json:"entropy"`
}

// ComputeEntropy computes the entropy of this string literal's value under the
// given character distribution, and sets its Entropy field to the resulting value.
func (s *String) ComputeEntropy(probs map[rune]float64) {
	s.Entropy = stringentropy.Calculate(s.Value, probs)
}

// LevenshteinDist computes the Levenshtein distance between the parsed and raw versions of
// this string literal. A character substitution is treated as deletion and insertion (2 operations).
func (s *String) LevenshteinDist() int {
	return levenshtein.DistanceForStrings([]rune(s.Raw), []rune(s.Value), levenshtein.DefaultOptions)
}

// Int records an integer literal occurring in source code. For languages without explicit
// integer types such as JavaScript, an Int literal is any numeric literal whose raw string
// representation in source code is parseable (with strconv.ParseInt) as an integer.
type Int struct {
	Value int64  `json:"value"`
	Raw   string `json:"raw"`
}

// Float records a floating point literal occurring in source code.
type Float struct {
	Value float64 `json:"value"`
	Raw   string  `json:"raw"`
}

// Comment records the entire text of a source code comment.
// It may contain newline characters.
type Comment struct {
	Text string `json:"text"`
}
