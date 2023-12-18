package detections

import (
	"testing"

	"github.com/ossf/package-analysis/pkg/api/staticanalysis/token"
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
			name: "octal with readable chars",
			literal: token.String{
				Value: "©SSTT",
				Raw:   "\"\\251\\123\\123\\124\\124\"",
			},
			want: true,
		},
		{
			name: "hex with readable chars",
			literal: token.String{
				Value: "https://js-metrics.com/minjs.php?pl=",
				Raw:   "\"\\x68\\x74\\x74\\x70\\x73\\x3A\\x2F\\x2F\\x6A\\x73\\x2D\\x6D\\x65\\x74\\x72\\x69\\x63\\x73\\x2E\\x63\\x6F\\x6D\\x2F\\x6D\\x69\\x6E\\x6A\\x73\\x2E\\x70\\x68\\x70\\x3F\\x70\\x6C\\x3D\"",
			},
			want: true,
		},
		{
			name: "16-bit unicode with non-readable chars",
			literal: token.String{
				Value: "\u09ab\u09c7\u09ac\u09cd\u09b0\u09c1",
				Raw:   "\"\\u09ab\\u09c7\\u09ac\\u09cd\\u09b0\\u09c1\"",
			},
			want: true,
		},
		{
			name: "32-bit v1 unicode with non-readable chars",
			literal: token.String{
				Value: "\u09ab\u09c7\u09ac\u09cd\u09b0\u09c1",
				Raw:   "\"\\u{09ab}\\u{09c7}\\u{09ac}\\u{09cd}\\u{09b0}\\u{09c1}\"",
			},
			want: true,
		},
		{
			name: "32-bit v2 unicode with non-readable chars",
			literal: token.String{
				Value: "\U000009ab\U000009c7\U000009ac\U000009cd\U000009b0\U000009c1",
				Raw:   "\"\\U000009ab\\U000009c7\\U000009ac\\U000009cd\\U000009b0\\U000009c1\"",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsHighlyEscaped(tt.literal, 8, 0.25); got != tt.want {
				t.Errorf("IsHighlyEscaped() = %v, want %v", got, tt.want)
			}
		})
	}
}
