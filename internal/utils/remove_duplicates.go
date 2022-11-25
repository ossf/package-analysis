package utils

/*
RemoveDuplicates takes a slice and returns a new slice with only the
unique elements from the input slice. Ordering of the elements in the
returned slice corresponds is done according to the earliest index
of each unique value in the input slice.
*/
func RemoveDuplicates[T comparable](items []T) []T {
	seenItems := make(map[T]struct{}) // empty structs take up no space
	var uniqueItems []T
	for _, item := range items {
		if _, seen := seenItems[item]; !seen {
			seenItems[item] = struct{}{}
			uniqueItems = append(uniqueItems, item)
		}
	}
	return uniqueItems
}
