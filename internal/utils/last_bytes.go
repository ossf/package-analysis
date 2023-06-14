package utils

// LastNBytes returns the last n bytes from b.
// If len(b) <= n, b itself is returned, otherwise a copy of the bytes is returned.
// If n is negative, the function will panic
func LastNBytes(b []byte, n int) []byte {
	if n < 0 {
		panic("n cannot be negative")
	}
	if len(b) <= n {
		return b
	}
	return b[(len(b) - n):]
}
