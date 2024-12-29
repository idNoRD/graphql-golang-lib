package token

import "strconv"

type Position struct {
	Start  int
	End    int
	Line   int
	Column int
}

type Pos int

func (p *Position) String() string {
	return strconv.Itoa(p.Line) + ":" + strconv.Itoa(p.Column) + " (" + strconv.Itoa(p.Start) + "-" + strconv.Itoa(p.End) + ")"
}
