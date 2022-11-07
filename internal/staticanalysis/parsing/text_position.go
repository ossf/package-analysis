package parsing

type TextPosition [2]int

func (pos TextPosition) Row() int {
	return pos[0]
}

func (pos TextPosition) Col() int {
	return pos[1]
}
