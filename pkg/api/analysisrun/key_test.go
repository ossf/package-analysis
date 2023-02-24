package analysisrun_test

import (
	"testing"

	"github.com/ossf/package-analysis/pkg/api/analysisrun"
	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

func TestStringify(t *testing.T) {
	tests := map[string]struct {
		input    analysisrun.Key
		expected string
	}{
		"simple stringify": {
			input:    analysisrun.Key{Name: "genericpackage", Version: "2.05.0", Ecosystem: pkgecosystem.NPM},
			expected: "npm-genericpackage-2.05.0",
		},
		"pkg name with space": {
			input:    analysisrun.Key{Name: "cool package", Version: "1.0.0", Ecosystem: pkgecosystem.PyPI},
			expected: "pypi-cool package-1.0.0",
		},
		"pkg name with forward slash": {
			input:    analysisrun.Key{Name: "@ada/evilpackage", Version: "99.0.0", Ecosystem: pkgecosystem.NPM},
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
