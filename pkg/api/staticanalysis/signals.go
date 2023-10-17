package staticanalysis

// EscapedString holds a string literal that contains a lot of character escaping.
// This may indicate obfuscation.
type EscapedString struct {
	Value           string `json:"value"`
	Raw             string `json:"raw"`
	LevenshteinDist int    `json:"levenshtein_dist"`
}

// SuspiciousIdentifier is an identifier that matches a specific rule intended
// to pick out (potentially) suspicious names. Name stores the actual identifier,
// and Rule holds the rule that the identifier matched against.
type SuspiciousIdentifier struct {
	Name string `json:"name"`
	Rule string `json:"rule"`
}
