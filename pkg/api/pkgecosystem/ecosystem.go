// The pkgecosystem package defines the open source ecosystems supported by Package Analysis.
package pkgecosystem

import (
	"errors"
	"fmt"
	"strings"
)

// Ecosystem represents an open source package ecosystem from which packages can be downloaded.
//
// It implements encoding.TextUnmarshaler and encoding.TextMarshaler so it can
// be used with flag.TextVar.
type Ecosystem string

const (
	None      Ecosystem = ""
	CratesIO  Ecosystem = "crates.io"
	NPM       Ecosystem = "npm"
	Packagist Ecosystem = "packagist"
	PyPI      Ecosystem = "pypi"
	RubyGems  Ecosystem = "rubygems"
)

// FlagUsage is a pregenerated string for use as usage text in flag.TextVar for
// an ecosystem flag.
const FlagUsage = "package ecosystem. Can be npm, pypi, rubygems, packagist, or crates.io"

// ErrUnsupported is returned by Ecosystem.UnmarshalText when bytes that do not
// correspond to a defined ecosystem constant is passed in as a parameter.
var ErrUnsupported = errors.New("ecosystem unsupported")

// EcosystemSet is a convenience type to interact with a collection of ecosystems.
type EcosystemSet []Ecosystem

// Lookup returns the Ecosystem in s that matches to text or None if there is no
// match. The bool returned will be true if there is a match, or false if there
// is no match.
func (s EcosystemSet) Lookup(text string) (Ecosystem, bool) {
	for _, e := range s {
		if text == string(e) {
			return e, true
		}
	}
	return None, false
}

// String implements the fmt.Stringer interface.
func (s EcosystemSet) String() string {
	var parts []string
	for _, e := range s {
		parts = append(parts, string(e))
	}
	return strings.Join(parts, ", ")
}

// SupportedEcosystems is a list of all the ecosystems supported.
var SupportedEcosystems = EcosystemSet{
	CratesIO,
	NPM,
	Packagist,
	PyPI,
	RubyGems,
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// It will only succeed when unmarshaling ecosytems in SupportedEcosystems or
// empty.
func (e *Ecosystem) UnmarshalText(text []byte) error {
	if s, ok := append(SupportedEcosystems, None).Lookup(string(text)); ok {
		*e = s
		return nil
	}
	return fmt.Errorf("%w: %s", ErrUnsupported, text)
}

// MarshalText implements the encoding.TextMarshaler interface.
func (e Ecosystem) MarshalText() ([]byte, error) {
	return []byte(e), nil
}

// String implements the fmt.Stringer interface.
func (e Ecosystem) String() string {
	return string(e)
}
