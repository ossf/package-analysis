// The pkgecosystem package defines the open source ecosystems supported by Package Analysis.
package pkgecosystem

import (
	"errors"
	"fmt"
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

// ErrUnsupported is returned by Ecosystem.UnmarshalText when bytes that do not
// correspond to a defined ecosystem constant is passed in as a parameter.
var ErrUnsupported = errors.New("ecosystem unsupported")

// SupportedEcosystems is a list of all the ecosystems supported.
var SupportedEcosystems = []Ecosystem{
	CratesIO,
	NPM,
	Packagist,
	PyPI,
	RubyGems,
}

// SupportedEcosystemsStrings is the list of supported ecosystems represented as
// strings.
var SupportedEcosystemsStrings = EcosystemsAsStrings(SupportedEcosystems)

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// It will only succeed when unmarshaling ecosytems in SupportedEcosystems or
// empty.
func (e *Ecosystem) UnmarshalText(text []byte) error {
	search := string(text)
	for _, s := range append(SupportedEcosystems, None) {
		if string(s) == search {
			*e = s
			return nil
		}
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

// EcosystemsAsStrings converts a slice of Ecosystems to a string slice.
func EcosystemsAsStrings(es []Ecosystem) []string {
	var s []string
	for _, e := range es {
		s = append(s, e.String())
	}
	return s
}
