package detections

import (
	"reflect"
	"testing"
)

const longBase64String = "IkxvcmVtIGlwc3VtIGRvbG9yIHNpdCBhbWV0LCBjb25zZWN0ZXR1ciBhZGlwaXNjaW5nIGVsaXQsIHNlZCBkby" +
	"BlaXVzbW9kIHRlbXBvciBpbmNpZGlkdW50IHV0IGxhYm9yZSBldCBkb2xvcmUgbWFnbmEgYWxpcXVhLiBVdCBlbmltIGFkIG1pbmltIHZlb" +
	"mlhbSwgcXVpcyBub3N0cnVkIGV4ZXJjaXRhdGlvbiB1bGxhbWNvIGxhYm9yaXMgbmlzaSB1dCBhbGlxdWlwIGV4IGVhIGNvbW1vZG8gY29u" +
	"c2VxdWF0LiBEdWlzIGF1dGUgaXJ1cmUgZG9sb3IgaW4gcmVwcmVoZW5kZXJpdCBpbiB2b2x1cHRhdGUgdmVsaXQgZXNzZSBjaWxsdW0gZG9" +
	"sb3JlIGV1IGZ1Z2lhdCBudWxsYSBwYXJpYXR1ci4gRXhjZXB0ZXVyIHNpbnQgb2NjYWVjYXQgY3VwaWRhdGF0IG5vbiBwcm9pZGVudCwgc3V" +
	"udCBpbiBjdWxwYSBxdWkgb2ZmaWNpYSBkZXNlcnVudCBtb2xsaXQgYW5pbSBpZCBlc3QgbGFib3J1bS4i"

func TestFindBase64Substrings(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		output []string
	}{
		{"empty", "", []string{}},
		{"16 lowercase chars", "abcdefghijklmnop", []string{}},
		{"16 uppercase chars", "ABCDEFGHIJKLMNOP", []string{}},
		{"16 digits", "1234123412341234", []string{}},
		{"16 chars lowercase hex", "0x0123456789abcd", []string{}},
		{"16 chars uppercase hex", "0XABCDEF12345678", []string{}},
		{"actual base64 no padding", "dGhpcyBpcyBhbiBvcmFuZ2UK", []string{"dGhpcyBpcyBhbiBvcmFuZ2UK"}},
		{"actual base64 1 padding", "dGhpcyBpcyBhIHBlYXI=", []string{"dGhpcyBpcyBhIHBlYXI="}},
		{"actual base64 2 padding", "dGhpcyBpcyBhbiBhcHBsZQ==", []string{"dGhpcyBpcyBhbiBhcHBsZQ=="}},
		{"actual base64 3 padding", "0XABCDEF12345678", []string{}},
		{"long base64 string", longBase64String, []string{longBase64String}},
		{
			"multiple base64 strings", longBase64String + " " + longBase64String,
			[]string{longBase64String, longBase64String},
		},
		{
			"multiple base64 strings 2", longBase64String + "!!!!====!!" + longBase64String,
			[]string{longBase64String, longBase64String},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindBase64Substrings(tt.input); !reflect.DeepEqual(got, tt.output) {
				t.Errorf("FindBase64Substrings() = %v, want %v", got, tt.output)
			}
		})
	}
}
