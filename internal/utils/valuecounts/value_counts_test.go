package valuecounts

import (
	"reflect"
	"testing"

	"github.com/ossf/package-analysis/internal/utils"
)

func TestCountData_ToValueCountPairs(t *testing.T) {
	tests := []struct {
		name string
		vc   ValueCounts
		want []Pair
	}{
		{
			"nil",
			nil,
			[]Pair{},
		},
		{
			"empty",
			ValueCounts{},
			[]Pair{},
		},
		{
			"single item",
			ValueCounts{0: 1},
			[]Pair{{0, 1}},
		},
		{
			"multiple items",
			ValueCounts{0: 1, 1: 2, 2: 3},
			[]Pair{{0, 1}, {1, 2}, {2, 3}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.vc.ToPairs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToPairs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromValueCountPairs(t *testing.T) {
	tests := []struct {
		name    string
		pairs   []Pair
		want    ValueCounts
		wantErr bool
	}{
		{
			"nil",
			nil,
			ValueCounts{},
			false,
		},
		{
			"empty non-nil",
			[]Pair{},
			ValueCounts{},
			false,
		},
		{
			"single item",
			[]Pair{{0, 1}},
			ValueCounts{0: 1},
			false,
		},
		{
			"multiple items",
			[]Pair{{0, 1}, {1, 2}},
			ValueCounts{0: 1, 1: 2},
			false,
		},
		{
			"repeated items",
			[]Pair{{0, 1}, {0, 1}},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromPairs(tt.pairs)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromPairs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromPairs() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCountData_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		vc      ValueCounts
		want    string
		wantErr bool
	}{
		{
			"nil",
			nil,
			"[]",
			false,
		},
		{
			"empty",
			ValueCounts{},
			"[]",
			false,
		},
		{
			"single item",
			ValueCounts{0: 1},
			`[ {"value": 0, "count": 1} ]`,
			false,
		},
		{
			"multiple items",
			ValueCounts{0: 1, 1: 2, 2: 3},
			`[ {"value":0, "count": 1}, {"value": 1, "count": 2}, {"value": 2, "count": 3} ]`,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBytes, err := tt.vc.MarshalJSON()
			got := string(gotBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if equal, err := utils.JSONEquals(gotBytes, []byte(tt.want)); err != nil {
				t.Errorf("MarshalJSON() error decoding JSON: %v", err)
			} else if !equal {
				t.Errorf("MarshalJSON() got = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestCountData_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		want    ValueCounts
		wantErr bool
	}{
		{
			"null",
			"null",
			ValueCounts{},
			false,
		},
		{
			"empty",
			"[]",
			ValueCounts{},
			false,
		},
		{
			"single item",
			`[{"value": 0, "count": 1}]`,
			ValueCounts{0: 1},
			false,
		},
		{
			"multiple items",
			`[{"value":0,"count":1},{"value":1,"count":2},{"value":2,"count":3}]`,
			ValueCounts{0: 1, 1: 2, 2: 3},
			false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valueCounts := New()

			if err := valueCounts.UnmarshalJSON([]byte(tt.json)); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(valueCounts, tt.want) {
				t.Errorf("UnmarshalJSON(): got %v, want %v", valueCounts, tt.want)
			}
		})
	}
}
