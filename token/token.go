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
	STRING_VALUE // TODO: remove
	STRING
	BLOCK_STRING
	COMMENT
)

var types = [...]string{
	EOF: "EOF",

	BANG:   "BANG",
	DOLLAR: "DOLLAR",
	AMP:    "AMP",
	LPAREN: "LPAREN",
	RPAREN: "RPAREN",
	SPREAD: "SPREAD",
	COLON:  "COLON",
	EQUALS: "EQUALS",
	AT:     "AT",
	LBRACK: "LBRACK",
	RBRACK: "RBRACK",
	LBRACE: "LBRACE",
	PIPE:   "PIPE",
	RBRACE: "RBRACE",

	NAME:         "NAME",
	INT:          "INT",
	FLOAT:        "FLOAT",
	STRING_VALUE: "STRING_VALUE", // TODO: remove
	STRING:       "STRING",
	BLOCK_STRING: "BLOCK_STRING",
	COMMENT:      "COMMENT",
}

func (t Type) String() string {
	return types[t]
}

// Token represents a lexical token.
type Token struct {
	Type    Type
	Literal string
	Start   int
	End     int
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%s) at %d-%d", types[t.Type], t.Literal, t.Start, t.End)
}
