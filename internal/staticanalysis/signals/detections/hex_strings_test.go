package detections

import (
	"reflect"
	"testing"
)

func TestFindHexSubstrings(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "empty",
			input: "",
			want:  nil,
		},
		{
			name:  "not hex",
			input: "abcdefghijklmnop",
			want:  nil,
		},
		{
			name:  "single hex",
			input: "abcdefabcdef12344",
			want:  []string{"abcdefabcdef12344"},
		},
		{
			name:  "two hex",
			input: "abcdefabcdef12344, 09acb8921308bac4",
			want:  []string{"abcdefabcdef12344", "09acb8921308bac4"},
		},
		{
			name:  "hex with prefix and non-hex suffix",
			input: "0xabcdefabcdef1234b09acb8921308bac4@02345",
			want:  []string{"abcdefabcdef1234b09acb8921308bac4"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindHexSubstrings(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindHexSubstrings() = %v, want %v", got, tt.want)
			}
		})
	}
}
