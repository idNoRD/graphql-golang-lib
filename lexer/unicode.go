package lexer

import (
	"errors"
	"fmt"
	"unicode/utf8"
)

func (l *Lexer) readEscapedUnicode() (rune, error) {
	l.readChar()

	if l.ch == '{' {
		return l.readUnicodeVariableWidth()
	}
	return l.readUnicodeFixedWidthOrSurrogate()
}

func (l *Lexer) readUnicodeFixedWidthOrSurrogate() (rune, error) {
	value, err := l.readUnicodeFixedWidth()
	if err != nil {
		return utf8.RuneError, err
	}
	if isUnicodeScalarValue(value) {
		return value, nil
	}
	if isLeadingSurrogate(value) {
		l.readChar()
		if l.ch != '\\' {
			return utf8.RuneError, l.newLexError(errors.New("expected '\\u' for trailing surrogate in Unicode escape sequence"))
		}
		l.readChar()
		if l.ch != 'u' {
			return utf8.RuneError, l.newLexError(errors.New("expected 'u' after '\\' in Unicode escape sequence"))
		}
		l.readChar()
		trailingValue, err := l.readUnicodeFixedWidth()
		if err != nil {
			return utf8.RuneError, err
		}
		if isTrailingSurrogate(trailingValue) {
			return combineSurrogates(value, trailingValue), nil
		}
		return utf8.RuneError, l.newLexError(fmt.Errorf("invalid trailing surrogate in Unicode escape sequence '%s'", l.getCapturedSequence()))
	}

	return utf8.RuneError, l.newLexError(fmt.Errorf("invalid Unicode escape sequence '%s'", l.getCapturedSequence()))
}

func (l *Lexer) readUnicodeFixedWidth() (rune, error) {
	var value rune
	for i := 0; i < 4; i++ {
		digit := hexDigitToInt(l.ch)
		if digit < 0 {
			return utf8.RuneError, l.newLexError(fmt.Errorf("invalid hex digit '%c' in Unicode escape sequence '%s'", l.ch, l.getCapturedSequence()))
		}
		value = (value << 4) | rune(digit)
		if i < 3 {
			l.readChar()
		}
	}
	return value, nil
}

func (l *Lexer) readUnicodeVariableWidth() (rune, error) {
	var value rune
	count := 0
	for {
		l.readChar()
		if l.ch == '}' {
			break
		}
		if l.ch == eof {
			return utf8.RuneError, l.newLexError(errors.New("unterminated Unicode escape sequence"))
		}
		digit := hexDigitToInt(l.ch)
		if digit < 0 {
			return utf8.RuneError, l.newLexError(fmt.Errorf("invalid hex digit '%c' in Unicode escape sequence '%s'", l.ch, l.getCapturedSequence()))
		}
		value = (value << 4) | rune(digit)
		count++
		if count > 8 {
			return utf8.RuneError, l.newLexError(fmt.Errorf("unicode escape sequence '%s' is too long", l.getCapturedSequence()))
		}
	}
	if count == 0 {
		return 0, l.newLexError(errors.New("unicode escape sequence cannot be empty"))
	}
	if !isUnicodeScalarValue(value) {
		return value, l.newLexError(fmt.Errorf("unicode escape sequence '%s' is out of range or invalid", l.getCapturedSequence()))
	}
	return value, nil
}
