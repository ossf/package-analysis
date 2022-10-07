package utils

func Transform[T, R any](ts []T, fn func(T) R) []R {
	result := make([]R, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func TransformMap[K comparable, V, R any](m map[K]V, fn func(K, V) R) []R {
	result := make([]R, len(m))
	i := 0
	for k, v := range m {
		result[i] = fn(k, v)
		i += 1
	}
	return result
}
