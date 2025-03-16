package utils

import (
	"math"
	"testing"
)

func TestFloatEquals(t *testing.T) {
	tests := []struct {
		name     string
		x1       float64
		x2       float64
		absTol   float64
		expected bool
	}{
		{name: "equal", x1: 1.0, x2: 1.0, absTol: 0.0, expected: true},
		{name: "gt tolerance", x1: 1.0, x2: 1.1, absTol: 0.1, expected: false},
		{name: "lt 1 OoM tolerance", x1: 1.0, x2: 1.1, absTol: 0.2, expected: true},
		{name: "lt 2 OoM tolerance", x1: 1.0, x2: 1.01, absTol: 0.02, expected: true},
		{name: "NaN always equal", x1: math.NaN(), x2: math.NaN(), absTol: 0.123456789, expected: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := FloatEquals(test.x1, test.x2, test.absTol)
			if actual != test.expected {
				t.Errorf(
					"FloatEquals(%f, %f, %f) expected to be %t but was %t",
					test.x1,
					test.x2,
					test.absTol,
					test.expected,
					actual,
				)
			}
		})
	}
}

func TestJSONEquals(t *testing.T) {
	tests := []struct {
		name     string
		j1       []byte
		j2       []byte
		expected bool
	}{
		{
			name:     "eq empty",
			j1:       []byte("{}"),
			j2:       []byte("{}"),
			expected: true,
		},
		{
			name:     "eq one key",
			j1:       []byte(`{"a": 1}`),
			j2:       []byte(`{"a": 1}`),
			expected: true,
		},
		{
			name:     "eq two keys",
			j1:       []byte(`{"a": 1, "b": 1}`),
			j2:       []byte(`{"a": 1, "b": 1}`),
			expected: true,
		},
		{
			name:     "eq two keys diff order",
			j1:       []byte(`{"a": 1, "b": 1}`),
			j2:       []byte(`{"b": 1, "a": 1}`),
			expected: true,
		},
		{
			name:     "deep eq",
			j1:       []byte(`{"a": {"b": 1}}`),
			j2:       []byte(`{"a": {"b": 1}}`),
			expected: true,
		},
		{
			name:     "deep eq diff order",
			j1:       []byte(`{"a": {"b": 1, "c": 1}}`),
			j2:       []byte(`{"a": {"c": 1, "b": 1}}`),
			expected: true,
		},
		{
			name:     "neq",
			j1:       []byte(`{"a": 1}`),
			j2:       []byte(`{"b": 1}`),
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := JSONEquals(test.j1, test.j2)

			if err != nil {
				t.Fatalf("JSONEquals() = %v; want no error", err)
			}

			if actual != test.expected {
				t.Errorf(
					"JSONEquals(%s, %s) expected to be %t but was %t",
					test.j1,
					test.j2,
					test.expected,
					actual,
				)
			}
		})
	}
}
