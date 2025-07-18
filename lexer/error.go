package lexer

import (
	"errors"
	"fmt"
)

type LexError struct {
	Column int
	Line   int
	Err    error
}

func (l *Lexer) newLexError(err error) error {
	var cur cursor
	if l.savedCursor == (cursor{}) {
		cur = l.cursor
	} else {
		cur = l.savedCursor
	}
	return &LexError{
		Line:   cur.line,
		Column: cur.column,
		Err:    err,
	}
}

func (e *LexError) Error() string {
	return fmt.Sprintf("Error at %d:%d: %s", e.Line, e.Column, e.Err.Error())
}

func (e *LexError) Is(target error) bool {
	var t *LexError
	ok := errors.As(target, &t)
	if !ok {
		return false
	}
	return e.Line == t.Line && e.Column == t.Column && e.Err.Error() == t.Err.Error()
}
