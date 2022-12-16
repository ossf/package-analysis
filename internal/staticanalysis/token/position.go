package token

type Position [2]int

func (pos Position) Row() int {
	return pos[0]
}

func (pos Position) Col() int {
	return pos[1]
}
