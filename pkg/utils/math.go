package utils

type Position struct {
	Offset int
	Length int
}

func (p Position) EndOffset() int {
	return p.Offset + p.Length
}

type Positions []Position

func (p Positions) Offset() int {
	if len(p) == 0 {
		return 0
	}
	return p[0].Offset
}

func (p Positions) EndOffset() int {
	l := len(p)
	if l == 0 {
		return 0
	}
	return p[l-1].EndOffset()
}
