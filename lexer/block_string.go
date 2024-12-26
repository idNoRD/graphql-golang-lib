package lexer

import (
	"errors"
	"strings"
	"unicode/utf8"
)

// https://spec.graphql.org/draft/#BlockString
func (l *Lexer) readBlockString() (string, error) {
	l.readChar() // consume first "
	l.readChar() // consume second "
	l.readChar() // consume third "

	var rawValue strings.Builder

	for {
		if l.ch == eof {
			return "", l.newLexError(errors.New("unterminated block string"))
		}
		//if l.ch < 0x20 {
		//	return "", l.newLexError(fmt.Errorf("invalid character in block string literal: '\\u%04X'", l.ch))
		//}
		if l.ch == '"' && l.peekChar() == '"' && l.peekCharAt(1) == '"' {
			l.readChar() // consume first "
			l.readChar() // consume second "
			l.readChar() // consume third "
			break
		}

		if l.ch == '\\' && l.peekChar() == '"' && l.peekCharAt(1) == '"' && l.peekCharAt(2) == '"' {
			l.readChar() // consume \
			l.readChar() // consume first "
			l.readChar() // consume second "
			l.readChar() // consume third "
			rawValue.WriteString(`"""`)
			continue
		}

		// Directly write ASCII characters as bytes
		if l.ch < utf8.RuneSelf {
			rawValue.WriteByte(byte(l.ch))
		} else {
			rawValue.WriteRune(l.ch)
		}
		l.readChar()
	}

	return l.processBlockStringValue(rawValue.String())
}

func (l *Lexer) processBlockStringValue(rawValue string) (string, error) {
	lines := splitLinesByLineTerminator(rawValue)

	// Determine common indentation (excluding the first line)
	var commonIndent = -1
	for i, line := range lines {
		if i == 0 {
			continue // Skip the first line
		}
		length := len(line)
		indent := leadingWhitespaceCount(line)
		if indent < length { // Non-empty line
			if commonIndent == -1 || indent < commonIndent {
				commonIndent = indent
			}
		}
	}

	// Remove common indentation from all lines except first
	if commonIndent > 0 {
		for i := 1; i < len(lines); i++ {
			if len(lines[i]) >= commonIndent {
				lines[i] = lines[i][commonIndent:]
			}
		}
	}

	lines = removeLeadingAndTrailingBlankLines(lines)

	if len(lines) == 0 {
		return "", nil // All lines are blank
	}

	// Reassemble the lines using \n
	formatted := strings.Join(lines, "\n")

	return formatted, nil
}

func splitLinesByLineTerminator(s string) []string {
	var lines []string
	start := 0

	for i := 0; i < len(s); i++ {
		if s[i] == '\n' || s[i] == '\r' {
			if i > start {
				lines = append(lines, s[start:i])
			} else {
				lines = append(lines, "")
			}
			// Handle '\r\n' as single delimiter
			if s[i] == '\r' && i+1 < len(s) && s[i+1] == '\n' {
				i++
			}
			start = i + 1
		}
	}

	// Append last line if there is remaining content after last delimiter
	if start < len(s) {
		lines = append(lines, s[start:])
	}

	return lines
}
