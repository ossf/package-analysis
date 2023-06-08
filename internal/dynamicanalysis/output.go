package dynamicanalysis

// lastNBytes returns the last n bytes from b.
// If len(b) < n, b itself is returned, otherwise a copy of the bytes is returned.
func lastNBytes(b []byte, n int) []byte {
	if len(b) < n {
		return b
	}
	return b[(len(b) - n):]
}
