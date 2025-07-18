package token

import "strconv"

type Position struct {
	Offset int
	Line   int
	Column int
}

func (p *Position) String() string {
	return strconv.Itoa(p.Line) + ":" + strconv.Itoa(p.Column) + " (" + strconv.Itoa(p.Offset) + ")"
}
