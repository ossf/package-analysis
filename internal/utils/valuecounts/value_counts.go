package valuecounts

import (
	"encoding/json"
	"fmt"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// ValueCounts stores unordered counts of integer values as a map
// from value (int) to count (int). It can be serialized to JSON
// as an array of (value, count) pairs.
type ValueCounts map[int]int

// Aside: I know using 'value' to refer to map keys is not great, but the
// other names I came up with like 'size' and 'length' were all usage-specific.

// Pair stores a single value and associated count pair
type Pair struct {
	Value int `json:"value"`
	Count int `json:"count"`
}

func New() ValueCounts {
	return ValueCounts{}
}

// ToPairs converts this ValueCounts into a list of (value, count) pairs.
// The values are sorted in increasing order so that the output is deterministic.
// If this ValueCounts is empty, returns an empty slice.
func (vc ValueCounts) ToPairs() []Pair {
	pairs := make([]Pair, 0, len(vc))

	// sort the values so that the output is in a deterministic order
	values := maps.Keys(vc)
	slices.Sort(values)

	for _, value := range values {
		count := vc[value]
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
		if _, seen := valueCounts[item.Value]; seen {
			return nil, fmt.Errorf("value occurs multiple times: %d", item.Value)
		}
		valueCounts[item.Value] = item.Count
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
