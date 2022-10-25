package dynamicanalysis

// lastLines returns a byte array containing only the last maxLines of lines
// in b.
//
// A line is defined the by the occurance of a '\n' in the output.
//
// If there are fewer than maxLines then the entire intput byte array will be
// returned.
//
// No more than maxBytes will be in the returned byte array.
func lastLines(b []byte, maxLines, maxBytes int) []byte {
	if len(b) == 0 {
		return b
	}
	// only consider the last maxBytes of b, or all of b if maxBytes is bigger
	// than the size of b.
	startOff := len(b) - maxBytes
	if startOff < 0 {
		startOff = 0
	}
	b = b[startOff:]
	// count each newline from the end of the bufffer
	lines := 0
	lineOff := len(b) - 1
	for off := len(b) - 1; off > 0; off-- {
		if b[off] == '\n' {
			lines++
			lineOff = off
		}
		if lines > maxLines {
			return b[lineOff+1:]
		}
	}
	return b
}
