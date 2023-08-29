package token

import "github.com/ossf/package-analysis/internal/staticanalysis/obfuscation/stringentropy"

type Identifier struct {
	Name    string         `json:"name"`
	Type    IdentifierType `json:"type"`
	Entropy float64        `json:"entropy"`
}

// ComputeEntropy computes the entropy of this identifier under the given
// character distribution and sets its Entropy field to the result value
func (i *Identifier) ComputeEntropy(probs map[rune]float64) {
	i.Entropy = stringentropy.Calculate(i.Name, probs)
}

type String struct {
	Value   string  `json:"value"`
	Raw     string  `json:"raw"`
	Entropy float64 `json:"entropy"`
}

// ComputeEntropy computes the entropy of this string literal under the given
// character distribution and sets its Entropy field to the result value
func (s *String) ComputeEntropy(probs map[rune]float64) {
	s.Entropy = stringentropy.Calculate(s.Value, probs)
}

type Int struct {
	Value int64  `json:"value"`
	Raw   string `json:"raw"`
}

type Float struct {
	Value float64 `json:"value"`
	Raw   string  `json:"raw"`
}

type Comment struct {
	Text string `json:"text"`
}
