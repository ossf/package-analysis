package escaping

import (
	"testing"

	"github.com/ossf/package-analysis/internal/staticanalysis/token"
)

func TestIsHighlyEscaped(t *testing.T) {
	tests := []struct {
		name    string
		literal token.String
		want    bool
	}{
		{
			name:    "empty",
			literal: token.String{},
			want:    false,
		},
		{
			name: "non escaped",
			literal: token.String{
				Value: "the quick brown fox jumps over the lazy dog",
				Raw:   "the quick brown fox jumps over the lazy dog",
			},
			want: false,
		},
		{
			name: "escaped resolves to readable",
			literal: token.String{
				Value: "https://js-metrics.com/minjs.php?pl=",
				Raw:   "\"\\x68\\x74\\x74\\x70\\x73\\x3A\\x2F\\x2F\\x6A\\x73\\x2D\\x6D\\x65\\x74\\x72\\x69\\x63\\x73\\x2E\\x63\\x6F\\x6D\\x2F\\x6D\\x69\\x6E\\x6A\\x73\\x2E\\x70\\x68\\x70\\x3F\\x70\\x6C\\x3D\"",
			},
			want: true,
		},
		{
			name: "escaped resolves to non-readable",
			literal: token.String{
				Value: "\u09ab\u09c7\u09ac\u09cd\u09b0\u09c1",
				Raw:   "\"\\u09ab\\u09c7\\u09ac\\u09cd\\u09b0\\u09c1\"",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsHighlyEscaped(tt.literal); got != tt.want {
				t.Errorf("IsHighlyEscaped() = %v, want %v", got, tt.want)
			}
		})
	}
}
