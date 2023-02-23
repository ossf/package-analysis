package pkgidentifier

import (
	"testing"
)

func TestStringify(t *testing.T) {
	tests := map[string]struct {
		input    PkgIdentifier
		expected string
	}{
		"simple stringify": {
			input:    PkgIdentifier{Name: "genericpackage", Version: "2.05.0", Ecosystem: "npm"},
			expected: "npm-genericpackage-2.05.0",
		},
		"pkg name with space": {
			input:    PkgIdentifier{Name: "cool package", Version: "1.0.0", Ecosystem: "pypi"},
			expected: "pypi-cool package-1.0.0",
		},
		"pkg name with forward slash": {
			input:    PkgIdentifier{Name: "@ada/evilpackage", Version: "99.0.0", Ecosystem: "npm"},
			expected: "npm-@ada/evilpackage-99.0.0",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.input.String()
			expected := test.expected
			if got != expected {
				t.Fatalf("%v: returned %v; expected %v", name, got, expected)
			}
		})
	}
}
