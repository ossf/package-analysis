package valuecounts

import (
	"reflect"
	"testing"

	"github.com/ossf/package-analysis/internal/utils"
)

func TestValueCounts_ToValueCountPairs(t *testing.T) {
	tests := []struct {
		name string
		vc   ValueCounts
		want []Pair
	}{
		{
			"nil",
			New(),
			[]Pair{},
		},
		{
			"empty",
			New(),
			[]Pair{},
		},
		{
			"single item",
			FromMap(map[int]int{0: 1}),
			[]Pair{{0, 1}},
		},
		{
			"multiple items",
			FromMap(map[int]int{0: 1, 1: 2, 2: 3}),
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
			New(),
			false,
		},
		{
			"empty non-nil",
			[]Pair{},
			New(),
			false,
		},
		{
			"single item",
			[]Pair{{0, 1}},
			FromMap(map[int]int{0: 1}),
			false,
		},
		{
			"multiple items",
			[]Pair{{0, 1}, {1, 2}},
			FromMap(map[int]int{0: 1, 1: 2}),
			false,
		},
		{
			"repeated items",
			[]Pair{{0, 1}, {0, 1}},
			New(),
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
			if err != nil {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromPairs() got %v, want %v", got, tt.want)
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
			ValueCounts{},
			"[]",
			false,
		},
		{
			"empty",
			New(),
			"[]",
			false,
		},
		{
			"single item",
			FromMap(map[int]int{0: 1}),
			`[ {"value": 0, "count": 1} ]`,
			false,
		},
		{
			"multiple items",
			FromMap(map[int]int{0: 1, 1: 2, 2: 3}),
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
				t.Errorf("MarshalJSON() got %s, want %s", got, tt.want)
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
			New(),
			false,
		},
		{
			"empty",
			"[]",
			New(),
			false,
		},
		{
			"single item",
			`[{"value": 0, "count": 1}]`,
			FromMap(map[int]int{0: 1}),
			false,
		},
		{
			"multiple items",
			`[{"value":0,"count":1},{"value":1,"count":2},{"value":2,"count":3}]`,
			FromMap(map[int]int{0: 1, 1: 2, 2: 3}),
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

func TestFromMap(t *testing.T) {
	tests := []struct {
		name string
		data map[int]int
		want ValueCounts
	}{
		{
			"nil",
			nil,
			New(),
		},
		{
			"empty",
			map[int]int{},
			New(),
		},
		{
			"basic",
			map[int]int{-1: 210, 10: 102, 0: 34, 3: 0},
			ValueCounts{
				data: map[int]int{-1: 210, 0: 34, 3: 0, 10: 102},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromMap(tt.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name string
		data []int
		want ValueCounts
	}{
		{
			"nil",
			nil,
			New(),
		},
		{
			"empty",
			[]int{},
			New(),
		},
		{
			"single",
			[]int{1},
			FromMap(map[int]int{1: 1}),
		},
		{
			"multiple",
			[]int{1, 2, 3, 4, 3, 2, 1, 2},
			FromMap(map[int]int{1: 2, 2: 3, 3: 2, 4: 1}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Count(tt.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Count() = %v, want %v", got, tt.want)
			}
		})
	}
}
