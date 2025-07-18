package lexer

import (
	"errors"
	"fmt"
	"github.com/gqlhub/gqlhub-core/token"
	"strings"
	"unicode/utf8"
)

type Lexer struct {
	input string
	ch    rune // current char
	cursor
	savedCursor cursor
}

type cursor struct {
	offset   int // Current position in input (points to current char)
	rdOffset int // Current reading position in input (after current char)
	line     int // Current line number
	column   int // Current column number
}

const (
	eof = -1
)

func New(input string) *Lexer {
	l := &Lexer{
		input: input,
		cursor: cursor{
			line: 1,
		},
	}

	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.ch == '\r' {
		l.line++
		l.column = 0
		if l.rdOffset < len(l.input) && l.input[l.rdOffset] == '\n' {
			l.rdOffset++
		}
	} else if l.ch == '\n' {
		l.line++
		l.column = 0
	}

	if l.rdOffset < len(l.input) {
		l.offset = l.rdOffset
		b := l.input[l.rdOffset]
		if b < utf8.RuneSelf { // ASCII
			l.ch = rune(b)
			l.rdOffset++
		} else { // Non-ASCII
			var size int
			l.ch, size = utf8.DecodeRuneInString(l.input[l.rdOffset:])
			l.rdOffset += size
		}
	} else {
		l.ch = eof
		l.offset = len(l.input)
	}
	l.column++
}

func (l *Lexer) NextToken() (tok token.Token, err error) {
	l.skipInsignificantChars()

	l.savedCursor = cursor{}

	tok.Start = l.offset

	switch {
	case isNameStart(l.ch):
		tok.Type = token.NAME
		tok.Literal = l.readName()
	case isDigit(l.ch) || l.ch == '-':
		tok.Type, tok.Literal, err = l.readNumber()
		if err != nil {
			return
		}
	default:
		switch l.ch {
		case '!':
			tok.Type = token.BANG
			l.readChar()
		case '$':
			tok.Type = token.DOLLAR
			l.readChar()
		case '&':
			tok.Type = token.AMP
			l.readChar()
		case '(':
			tok.Type = token.LPAREN
			l.readChar()
		case ')':
			tok.Type = token.RPAREN
			l.readChar()
		case '.':
			ch := l.peekChar()
			switch {
			case ch == '.' && l.peekCharAt(1) == '.':
				l.readChar()
				l.readChar()
				l.readChar()
				tok.Type = token.SPREAD
			case isDigit(ch):
				err = l.newLexError(fmt.Errorf("invalid number, expected digit before '.'"))
				return
			default:
				err = l.newLexError(errors.New("unexpected '.'"))
				return
			}
		case ':':
			tok.Type = token.COLON
			l.readChar()
		case '=':
			tok.Type = token.EQUALS
			l.readChar()
		case '@':
			tok.Type = token.AT
			l.readChar()
		case '[':
			tok.Type = token.LBRACK
			l.readChar()
		case ']':
			tok.Type = token.RBRACK
			l.readChar()
		case '{':
			tok.Type = token.LBRACE
			l.readChar()
		case '|':
			tok.Type = token.PIPE
			l.readChar()
		case '}':
			tok.Type = token.RBRACE
			l.readChar()
		case '"':
			if l.peekChar() == '"' && l.peekCharAt(1) == '"' {
				tok.Type = token.BLOCK_STRING
				tok.Literal, err = l.readBlockString()
				if err != nil {
					return
				}
			} else {
				tok.Type = token.STRING
				tok.Literal, err = l.readString()
				if err != nil {
					return
				}
			}
		case '#':
			tok.Type = token.COMMENT
			tok.Literal = l.readComment()
		case eof:
			tok.Type = token.EOF
		default:
			err = l.newLexError(fmt.Errorf("unexpected character '%s'", printChar(l.ch)))
			return
		}
	}
	tok.End = l.offset
	return
}

func (l *Lexer) readName() string {
	start := l.offset
	for isNameContinue(l.ch) {
		l.readChar()
	}
	return l.input[start:l.offset]
}

func (l *Lexer) readNumber() (token.Type, string, error) {
	start := l.offset
	tokType := token.INT
	if l.ch == '-' {
		l.readChar()
	}
	if l.ch == '0' {
		l.readChar()
		if isDigit(l.ch) {
			return tokType, "", l.newLexError(fmt.Errorf("invalid number, unexpected digit after 0: '%s'", printChar(l.ch)))
		}
	} else {
		if !isDigit(l.ch) {
			return tokType, "", l.newLexError(fmt.Errorf("invalid number, expected digit but got '%s'", printChar(l.ch)))
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	if l.ch == '.' {
		tokType = token.FLOAT
		l.readChar()
		if !isDigit(l.ch) {
			return tokType, "", l.newLexError(fmt.Errorf("invalid number, expected digit but got '%s'", printChar(l.ch)))
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	if isExponentIndicator(l.ch) {
		tokType = token.FLOAT
		l.readChar()
		if isSign(l.ch) {
			l.readChar()
		}
		if !isDigit(l.ch) {
			return tokType, "", l.newLexError(fmt.Errorf("invalid number, expected digit but got '%s'", printChar(l.ch)))
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	// The numeric literals IntValue and FloatValue both restrict being immediately followed by a letter (or other NameStart)
	// https://spec.graphql.org/draft/#note-dea61
	if l.ch == '.' || isNameStart(l.ch) {
		return tokType, "", l.newLexError(fmt.Errorf("invalid number, expected digit but got '%c'", l.ch))
	}

	return tokType, l.input[start:l.offset], nil
}

// https://spec.graphql.org/draft/#StringValue
func (l *Lexer) readString() (string, error) {
	var result strings.Builder

	l.readChar() // consume "

	for l.ch != '"' {
		if l.ch == eof || isLineTerminator(l.ch) {
			return "", l.newLexError(errors.New("unterminated string"))
		}
		if l.ch < 0x20 {
			return "", l.newLexError(fmt.Errorf("invalid character in string literal: '\\u%04X'", l.ch))
		}
		if l.ch == '\\' { // EscapedCharacter and EscapedUnicode
			l.saveCurrentCursor()
			l.readChar()

			switch l.ch {
			case 'u':
				char, err := l.readEscapedUnicode()
				if err != nil {
					return "", err
				}
				result.WriteRune(char)
			default:
				if esc, ok := escapeChars[l.ch]; ok {
					result.WriteByte(esc)
				} else {
					return "", l.newLexError(fmt.Errorf("unknown escape sequence '\\%c'", l.ch))
				}
			}
		} else {
			result.WriteRune(l.ch)
		}
		l.readChar()
	}
	l.readChar() // consume closing "
	return result.String(), nil
}

func (l *Lexer) readComment() string {
	l.readChar()
	start := l.offset
	for !isLineTerminator(l.ch) && l.ch != eof {
		l.readChar()
	}
	return l.input[start:l.offset]
}

func (l *Lexer) skipInsignificantChars() {
	for isWhiteSpace(l.ch) || isLineTerminator(l.ch) || l.ch == ',' {
		l.readChar()
	}
}

func (l *Lexer) peekChar() rune {
	if l.rdOffset >= len(l.input) {
		return 0
	}
	ch, _ := decodeRuneAt(l.input[l.rdOffset:])
	return ch
}

func (l *Lexer) peekCharAt(offset int) rune {
	pos := l.rdOffset
	for i := 0; i < offset; i++ {
		if pos >= len(l.input) {
			return 0
		}
		_, w := decodeRuneAt(l.input[pos:])
		pos += w
	}
	if pos >= len(l.input) {
		return 0
	}
	ch, _ := decodeRuneAt(l.input[pos:])
	return ch
}

func (l *Lexer) saveCurrentCursor() {
	l.savedCursor = l.cursor
}

func (l *Lexer) getCapturedSequence() string {
	return l.input[l.savedCursor.offset:l.rdOffset]
}
