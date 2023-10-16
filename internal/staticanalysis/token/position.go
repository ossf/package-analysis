package token

// Position records the position of a source code token
// in terms of row and column in the original source file.
type Position [2]int

func (pos Position) Row() int {
	return pos[0]
}

func (pos Position) Col() int {
	return pos[1]
}
