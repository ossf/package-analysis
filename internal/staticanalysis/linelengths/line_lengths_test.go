package linelengths

import (
	"reflect"
	"testing"
)

func TestSourceStringLineLengths(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    []int
		wantErr bool
	}{
		{
			name: "test simple multiline",
			source: `
One
Two
Three
Four
Five
`,
			want:    []int{0, 3, 3, 5, 4, 4},
			wantErr: false,
		},
		{
			name:    "test simple single line",
			source:  `One Two Three Four Five`,
			want:    []int{23},
			wantErr: false,
		},
		{
			name:    "test empty string",
			source:  ``,
			want:    []int{0},
			wantErr: false,
		},
		{
			name:    "test single char",
			source:  "a",
			want:    []int{1},
			wantErr: false,
		},
		{
			name: "test empty newline",
			source: `
`,
			want:    []int{0},
			wantErr: false,
		},

		{
			name:    "test carriage return",
			source:  "\r\n",
			want:    []int{0},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetLineLengths("", tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLineLengths() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetLineLengths() got = %v, want %v", got, tt.want)
			}
		})
	}
}
