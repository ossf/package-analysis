package utils

import (
	"reflect"
	"testing"
)

type lastNBytesTestCase struct {
	name  string
	bytes []byte
	n     int
	want  []byte
}

func TestLastNBytes(t *testing.T) {
	tests := []lastNBytesTestCase{
		{
			"empty_0",
			[]byte{},
			0,
			[]byte{},
		},
		{
			"empty_10",
			[]byte{},
			10,
			[]byte{},
		},
		{
			"abcd_0",
			[]byte{'a', 'b', 'c', 'd'},
			0,
			[]byte{},
		},
		{
			"abcd_1",
			[]byte{'a', 'b', 'c', 'd'},
			1,
			[]byte{'d'},
		},
		{
			"abcd_4",
			[]byte{'a', 'b', 'c', 'd'},
			4,
			[]byte{'a', 'b', 'c', 'd'},
		},
		{
			"abcd_5",
			[]byte{'a', 'b', 'c', 'd'},
			5,
			[]byte{'a', 'b', 'c', 'd'},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LastNBytes(tt.bytes, tt.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LastNBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
