package token

import "fmt"

// Type represents the type of lexical tokens.
type Type int

const (
	EOF Type = iota

	BANG
	DOLLAR
	AMP
	LPAREN
	RPAREN
	SPREAD
	COLON
	EQUALS
	AT
	LBRACK
	RBRACK
	LBRACE
	PIPE
	RBRACE

	NAME
	INT
	FLOAT
	STRING_VALUE
	COMMENT
)

// Token represents a lexical token.
type Token struct {
	Type    Type
	Literal string
	Position
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%s) at %d:%d (%d-%d)", t.Type, t.Literal, t.Line, t.Column, t.Start, t.End)
}
