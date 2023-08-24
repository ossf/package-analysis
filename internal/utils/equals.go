package utils

import (
	"encoding/json"
	"math"
	"reflect"
)

// FloatEquals compares two floats and returns true if they are both
// within absTol of each other, or are both NaN.
// Note that normally NaN != NaN, but we define it as true because it's
// convenient for comparing arrays and structs that contain floats.
func FloatEquals(x1, x2, absTol float64) bool {
	return x1 == x2 || math.Abs(x1-x2) < absTol || (math.IsNaN(x1) && math.IsNaN(x2))
}

// JSONEquals compares two byte sequences containing JSON data and returns true if
// 1) both j1 and j2 contain valid JSON data, and
// 2) the JSON objects that they represent are equal.
// If j1 or j2 contain invalid JSON data, an error is returned.
func JSONEquals(j1, j2 []byte) (bool, error) {
	// Adapted from https://stackoverflow.com/a/32409106
	var o1, o2 interface{}
	if err := json.Unmarshal(j1, &o1); err != nil {
		return false, err
	}
	if err := json.Unmarshal(j2, &o2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(o1, o2), nil

}
