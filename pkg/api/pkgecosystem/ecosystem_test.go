package pkgecosystem_test

import (
	"bytes"
	"testing"

	"github.com/ossf/package-analysis/pkg/api/pkgecosystem"
)

func TestEcosystemMarshalText(t *testing.T) {
	tests := []struct {
		name string
		eco  pkgecosystem.Ecosystem
		want []byte
	}{
		{
			name: "npm",
			eco:  pkgecosystem.NPM,
			want: []byte("npm"),
		},
		{
			name: "unsupported",
			eco:  pkgecosystem.Ecosystem("this is a test"),
			want: []byte("this is a test"),
		},
		{
			name: "empty",
			eco:  pkgecosystem.None,
			want: []byte{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, _ := test.eco.MarshalText()
			if !bytes.Equal(got, test.want) {
				t.Errorf("MarshalText() = %v; want %v", got, test.want)
			}
		})
	}
}

func TestEcosystemUnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    pkgecosystem.Ecosystem
		wantErr bool
	}{
		{
			name:  "npm",
			input: []byte("npm"),
			want:  pkgecosystem.NPM,
		},
		{
			name:  "crates.io",
			input: []byte("crates.io"),
			want:  pkgecosystem.CratesIO,
		},
		{
			name:    "unsupported",
			input:   []byte("this is a test"),
			wantErr: true,
		},
		{
			name:  "empty",
			input: []byte{},
			want:  pkgecosystem.None,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var got pkgecosystem.Ecosystem
			err := got.UnmarshalText(test.input)
			if test.wantErr && err == nil {
				t.Fatal("UnmarshalText() is nil; want error")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("UnmarshalText() = %v; want nil", err)
			}
			if got != test.want {
				t.Errorf("UnmarshalText() parsed %v; want %v", got, test.want)
			}
		})
	}
}

func TestEcosystemString(t *testing.T) {
	tests := []struct {
		name string
		eco  pkgecosystem.Ecosystem
		want string
	}{
		{
			name: "npm",
			eco:  pkgecosystem.NPM,
			want: "npm",
		},
		{
			name: "unsupported",
			eco:  pkgecosystem.Ecosystem("this is a test"),
			want: "this is a test",
		},
		{
			name: "empty",
			eco:  pkgecosystem.Ecosystem(""),
			want: "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.eco.String()
			if got != test.want {
				t.Errorf("String() = %v; want %v", got, test.want)
			}
		})
	}
}

func TestEcosystemSetLookup(t *testing.T) {
	tests := []struct {
		name   string
		set    pkgecosystem.EcosystemSet
		input  string
		want   pkgecosystem.Ecosystem
		wantOk bool
	}{
		{
			name:   "empty set",
			set:    pkgecosystem.EcosystemSet{},
			input:  "npm",
			want:   pkgecosystem.None,
			wantOk: false,
		},
		{
			name:   "one entry",
			set:    pkgecosystem.EcosystemSet{pkgecosystem.NPM},
			input:  "npm",
			want:   pkgecosystem.NPM,
			wantOk: true,
		},
		{
			name: "not found",
			set: pkgecosystem.EcosystemSet{
				pkgecosystem.CratesIO,
				pkgecosystem.PyPI,
			},
			input:  "npm",
			want:   pkgecosystem.None,
			wantOk: false,
		},
		{
			name: "found",
			set: pkgecosystem.EcosystemSet{
				pkgecosystem.CratesIO,
				pkgecosystem.PyPI,
				pkgecosystem.NPM,
			},
			input:  "pypi",
			want:   pkgecosystem.PyPI,
			wantOk: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotOk := test.set.Lookup(test.input)
			if gotOk != test.wantOk {
				t.Errorf("Lookup() = %v; want %v", gotOk, test.wantOk)
			}
			if got != test.want {
				t.Errorf("Lookup() = %v; want %v", got, test.want)
			}
		})
	}
}

func TestEcosystemSetString(t *testing.T) {
	tests := []struct {
		name string
		set  pkgecosystem.EcosystemSet
		want string
	}{
		{
			name: "empty set",
			set:  pkgecosystem.EcosystemSet{},
			want: "",
		},
		{
			name: "one entry",
			set:  pkgecosystem.EcosystemSet{pkgecosystem.Ecosystem("one")},
			want: "one",
		},
		{
			name: "multiple entries",
			set: pkgecosystem.EcosystemSet{
				pkgecosystem.Ecosystem("one"),
				pkgecosystem.Ecosystem("two"),
				pkgecosystem.Ecosystem("three"),
			},
			want: "one, two, three",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.set.String()
			if got != test.want {
				t.Errorf("String() = %v; want %v", got, test.want)
			}
		})
	}
}
