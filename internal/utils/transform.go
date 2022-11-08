package utils

// Transform applies the given transform function fn: T -> R to each element t of slice ts
// and returns a slice containing the corresponding results.
func Transform[T, R any](ts []T, fn func(T) R) []R {
	result := make([]R, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}
