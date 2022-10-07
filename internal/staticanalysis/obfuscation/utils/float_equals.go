package utils

import "math"

func FloatEquals(x1, x2, absTol float64) bool {
	return x1 == x2 || math.Abs(x1-x2) < absTol || (math.IsNaN(x1) && math.IsNaN(x2))
}
