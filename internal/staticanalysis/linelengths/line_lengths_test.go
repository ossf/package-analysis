package linelengths

import (
	"reflect"
	"testing"
)

func TestSourceStringLineLengths(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		want    map[int]int
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
			want:    map[int]int{0: 1, 3: 2, 4: 2, 5: 1},
			wantErr: false,
		},
		{
			name:    "test simple single line",
			source:  `One Two Three Four Five`,
			want:    map[int]int{23: 1},
			wantErr: false,
		},
		{
			name:    "test empty string",
			source:  ``,
			want:    map[int]int{0: 1},
			wantErr: false,
		},
		{
			name:    "test single char",
			source:  "a",
			want:    map[int]int{1: 1},
			wantErr: false,
		},
		{
			name: "test empty newline",
			source: `
`,
			want:    map[int]int{0: 1},
			wantErr: false,
		},

		{
			name:    "test carriage return",
			source:  "\r\n",
			want:    map[int]int{0: 1},
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
