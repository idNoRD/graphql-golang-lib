package lexer

import (
	"fmt"
	"unicode/utf8"
)

func isLetter(ch rune) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}

func isDigit(ch rune) bool {
	return '0' <= ch && ch <= '9'
}

func isLineTerminator(ch rune) bool {
	return ch == '\n' /* Line Feed (LF) */ || ch == '\r' /* Carriage Return (CR) */
}

func isNameStart(ch rune) bool {
	return isLetter(ch) || ch == '_'
}

func isNameContinue(ch rune) bool {
	return isLetter(ch) || isDigit(ch) || ch == '_'
}

func isExponentIndicator(ch rune) bool {
	return ch == 'e' || ch == 'E'
}

func isSign(ch rune) bool {
	return ch == '+' || ch == '-'
}

func isHexDigit(ch rune) bool {
	return ('0' <= ch && ch <= '9') || ('a' <= ch && ch <= 'f') || ('A' <= ch && ch <= 'F')
}

func isUnicodeScalarValue(ch rune) bool {
	return (ch >= 0x0000 && ch <= 0xD7FF) || (ch >= 0xE000 && ch <= 0x10FFFF)
}

func isSurrogate(r rune) bool {
	return r >= 0xD800 && r <= 0xDFFF
}

func isLeadingSurrogate(ch rune) bool {
	return ch >= 0xD800 && ch <= 0xDBFF
}

func isTrailingSurrogate(ch rune) bool {
	return ch >= 0xDC00 && ch <= 0xDFFF
}

func combineSurrogates(leading, trailing rune) rune {
	return (leading-0xD800)*0x400 + (trailing - 0xDC00) + 0x10000
}

func decodeRuneAt(s string) (rune, int) {
	if len(s) < 1 { // TODO: do weed this?
		return utf8.RuneError, 0
	}

	ch := rune(s[0])
	if ch < utf8.RuneSelf {
		return ch, 1
	}

	return utf8.DecodeRuneInString(s)
}

func leadingWhitespaceCount(s string) int {
	count := 0
	for _, ch := range s {
		if isWhiteSpace(ch) {
			count++
		} else {
			break
		}
	}
	return count
}

func isBlank(s string) bool {
	for _, ch := range s {
		if !isWhiteSpace(ch) {
			return false
		}
	}
	return true
}

func isWhiteSpace(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

func printChar(ch rune) string {
	if ch == eof {
		return "<EOF>"
	}
	if ch >= 0x0020 && ch <= 0x007E {
		if ch == '"' {
			return `\"`
		}
		return fmt.Sprintf("%c", ch)
	}
	return fmt.Sprintf("U+%04X", ch)
}

func removeLeadingAndTrailingBlankLines(lines []string) []string {
	start, end := 0, len(lines)

	for start < end && isBlank(lines[start]) {
		start++
	}
	for end > start && isBlank(lines[end-1]) {
		end--
	}
	return lines[start:end]
}

func hexDigitToInt(ch rune) int {
	switch {
	case '0' <= ch && ch <= '9':
		return int(ch - '0')
	case 'a' <= ch && ch <= 'f':
		return int(ch-'a') + 10
	case 'A' <= ch && ch <= 'F':
		return int(ch-'A') + 10
	}
	return -1
}

var escapeChars = map[rune]byte{
	'"':  '"',
	'\\': '\\',
	'/':  '/',
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
}
