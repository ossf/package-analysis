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
	CratesIO  Ecosystem = "crates.io"
	NPM       Ecosystem = "npm"
	Packagist Ecosystem = "packagist"
	PyPI      Ecosystem = "pypi"
	RubyGems  Ecosystem = "rubygems"
)

var ErrUnsupported = errors.New("ecosystem unsupported")

// SupportedEcosystems is a list of all the ecosystems supported.
var SupportedEcosystems = []Ecosystem{
	CratesIO,
	NPM,
	Packagist,
	PyPI,
	RubyGems,
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
//
// It will only succeed when unmarshaling ecosytems in SupportedEcosystems.
func (e *Ecosystem) UnmarshalText(text []byte) error {
	candidate := string(text)
	for _, s := range SupportedEcosystems {
		if string(s) == candidate {
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
