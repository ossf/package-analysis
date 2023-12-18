package valuecounts

import (
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// ValueCounts stores unordered counts of integer values as a map
// from value (int) to count (int). It can be serialized to JSON
// as an array of (value, count) pairs.
type ValueCounts struct {
	data map[int]int
}

// Aside: using 'value' to refer to map keys is not great,
// but names like 'size' and 'length' are all usage-specific.

// Pair stores a single value and associated count pair
type Pair struct {
	Value int `json:"value"`
	Count int `json:"count"`
}

// New creates a new empty ValueCounts object
func New() ValueCounts {
	return ValueCounts{
		data: map[int]int{},
	}
}

// FromMap creates a new ValueCounts object and initialises its counts from the given map
func FromMap(data map[int]int) ValueCounts {
	vc := New()
	for value, count := range data {
		vc.data[value] = count
	}
	return vc
}

// Count produces a new ValueCounts by counting repetitions of values in the input data
func Count(data []int) ValueCounts {
	vc := New()
	for _, value := range data {
		vc.data[value] += 1
	}
	return vc
}

// Len returns the number of values stored by this ValueCounts.
// It is equivalent to the length of the slice returned by ToPairs()
func (vc ValueCounts) Len() int {
	return len(vc.data)
}

// String() returns a string representation of this ValueCounts
// with values sorted in ascending order
func (vc ValueCounts) String() string {
	pairStrings := make([]string, 0, len(vc.data))
	for _, pair := range vc.ToPairs() {
		pairStrings = append(pairStrings, fmt.Sprintf("%d: %d", pair.Value, pair.Count))
	}
	return "[" + strings.Join(pairStrings, ", ") + " ]"
}

// ToPairs converts this ValueCounts into a list of (value, count) pairs.
// The values are sorted in increasing order so that the output is deterministic.
// If this ValueCounts is empty, returns an empty slice.
func (vc ValueCounts) ToPairs() []Pair {
	pairs := make([]Pair, 0, len(vc.data))

	// sort the values so that the output is in a deterministic order
	values := maps.Keys(vc.data)
	slices.Sort(values)

	for _, value := range values {
		count := vc.data[value]
		pairs = append(pairs, Pair{Value: value, Count: count})
	}

	return pairs
}

// FromPairs converts a list of (value, count) pairs back into ValueCounts.
// If the same value occurs multiple times in the list, an error is raised.
// If pairs is nil or empty but non-nil, an empty ValueCounts is returned.
func FromPairs(pairs []Pair) (ValueCounts, error) {
	valueCounts := New()

	for _, item := range pairs {
		if _, seen := valueCounts.data[item.Value]; seen {
			return ValueCounts{}, fmt.Errorf("value occurs multiple times: %d", item.Value)
		}
		valueCounts.data[item.Value] = item.Count
	}

	return valueCounts, nil
}

// MarshalJSON serialises this ValueCounts into a JSON array of {value, count} pairs.
func (vc ValueCounts) MarshalJSON() ([]byte, error) {
	return json.Marshal(vc.ToPairs())
}

/*
UnmarshalJSON converts a JSON-serialised ValueCounts object, serialised
using MarshalJSON, back into a Go object. Existing counts are discarded.

Note that the serialised data is formatted as an array of (value, count) pairs,
and it is possible to create an array with multiple counts for the same value.
If this occurs, an error will be raised.

If any error is encountered, the original countData object is not modified.
*/
func (vc *ValueCounts) UnmarshalJSON(data []byte) error {
	var pair []Pair

	if err := json.Unmarshal(data, &pair); err != nil {
		return err
	}

	valueCounts, err := FromPairs(pair)
	if err != nil {
		return err
	}

	*vc = valueCounts
	return nil
}
