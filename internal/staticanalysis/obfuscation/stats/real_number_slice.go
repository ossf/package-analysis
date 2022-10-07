package stats

// RealNumberSlice
// Defines interface type needed for sorting of generic RealNumber slices
type RealNumberSlice[T RealNumber] []T

func (slice RealNumberSlice[T]) Len() int {
	return len(slice)
}

func (slice RealNumberSlice[T]) Less(i, j int) bool {
	return slice[i] < slice[j]
}

func (slice RealNumberSlice[T]) Swap(i, j int) {
	temp := slice[i]
	slice[i] = slice[j]
	slice[j] = temp
}
