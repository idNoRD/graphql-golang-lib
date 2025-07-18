package lexer

import (
	"errors"
	"fmt"
	"github.com/gqlhub/gqlhub-core/token"
	"reflect"
	"testing"
)

func TestNextToken_Punctuator(t *testing.T) {
	runValidTests(t, []struct {
		name          string
		input         string
		expectedToken token.Token
	}{
		{"Bang", "!", token.Token{Type: token.BANG, Start: 0, End: 1}},
		{"Dollar", "$", token.Token{Type: token.DOLLAR, Start: 0, End: 1}},
		{"Ampersand", "&", token.Token{Type: token.AMP, Start: 0, End: 1}},
		{"Left Parenthesis", "(", token.Token{Type: token.LPAREN, Start: 0, End: 1}},
		{"Right Parenthesis", ")", token.Token{Type: token.RPAREN, Start: 0, End: 1}},
		{"Colon", ":", token.Token{Type: token.COLON, Start: 0, End: 1}},
		{"Equals", "=", token.Token{Type: token.EQUALS, Start: 0, End: 1}},
		{"At", "@", token.Token{Type: token.AT, Start: 0, End: 1}},
		{"Left Bracket", "[", token.Token{Type: token.LBRACK, Start: 0, End: 1}},
		{"Right Bracket", "]", token.Token{Type: token.RBRACK, Start: 0, End: 1}},
		{"Left Brace", "{", token.Token{Type: token.LBRACE, Start: 0, End: 1}},
		{"Pipe", "|", token.Token{Type: token.PIPE, Start: 0, End: 1}},
		{"Right Brace", "}", token.Token{Type: token.RBRACE, Start: 0, End: 1}},
	})
}

func TestNextToken_LineTerminators(t *testing.T) {
	runValidTests(t, []struct {
		name          string
		input         string
		expectedToken token.Token
	}{
		{"Skip newline", "\nhello", token.Token{Type: token.NAME, Literal: "hello", Start: 1, End: 6}},
		{"Skip carriage return", "\rhello", token.Token{Type: token.NAME, Literal: "hello", Start: 1, End: 6}},
		{"Skip carriage return and newline", "\r\nhello", token.Token{Type: token.NAME, Literal: "hello", Start: 2, End: 7}},
		{"Skip newline and carriage return", "\n\rhello", token.Token{Type: token.NAME, Literal: "hello", Start: 2, End: 7}},
		{"Skip CR, CRLF, and LF", "\r\r\n\nhello", token.Token{Type: token.NAME, Literal: "hello", Start: 4, End: 9}},
		{"Skip LF, LFCR, and CR", "\n\n\r\rhello", token.Token{Type: token.NAME, Literal: `hello`, Start: 4, End: 9}},
	})
}

func TestNextToken_LinesAndColumns(t *testing.T) {
	runValidTests(t, []struct {
		name          string
		input         string
		expectedToken token.Token
	}{
		{"Track lines and columns with mixed line breaks and spaces", "\r \r\n \n   hello", token.Token{Type: token.NAME, Literal: "hello", Start: 9, End: 14}},
	})
}

func TestNextToken_Whitespace(t *testing.T) {
	runValidTests(t, []struct {
		name          string
		input         string
		expectedToken token.Token
	}{
		{"Skip horizontal tabs", "\t\thello\t", token.Token{Type: token.NAME, Literal: "hello", Start: 2, End: 7}},
		{"Skip leading spaces", "    hello", token.Token{Type: token.NAME, Literal: "hello", Start: 4, End: 9}},
	})
}

func TestNextToken_ValidNumbers(t *testing.T) {
	runValidTests(t, []struct {
		name          string
		input         string
		expectedToken token.Token
	}{
		// Integers
		{"Integer zero", "0", token.Token{Type: token.INT, Literal: "0", Start: 0, End: 1}},
		{"Integer", "3", token.Token{Type: token.INT, Literal: "3", Start: 0, End: 1}},
		{"Negative integer", "-3", token.Token{Type: token.INT, Literal: "-3", Start: 0, End: 2}},

		// Floats without Exponent
		{"Float", "3.1415", token.Token{Type: token.FLOAT, Literal: "3.1415", Start: 0, End: 6}},
		{"Float with leading zero", "0.123", token.Token{Type: token.FLOAT, Literal: "0.123", Start: 0, End: 5}},

		// Negative floats without Exponent
		{"Negative float", "-3.1415", token.Token{Type: token.FLOAT, Literal: "-3.1415", Start: 0, End: 7}},
		{"Negative float with leading zero", "-0.123", token.Token{Type: token.FLOAT, Literal: "-0.123", Start: 0, End: 6}},

		// Floats with Exponent
		{"Float with lowercase exponent", "12345e3", token.Token{Type: token.FLOAT, Literal: "12345e3", Start: 0, End: 7}},
		{"Float with uppercase exponent", "12345E3", token.Token{Type: token.FLOAT, Literal: "12345E3", Start: 0, End: 7}},
		{"Float with positive exponent sign", "12345e+3", token.Token{Type: token.FLOAT, Literal: "12345e+3", Start: 0, End: 8}},
		{"Float with negative exponent sign", "12345e-3", token.Token{Type: token.FLOAT, Literal: "12345e-3", Start: 0, End: 8}},
		{"Float with zero exponent", "12345e0", token.Token{Type: token.FLOAT, Literal: "12345e0", Start: 0, End: 7}},
		{"Float with large exponent", "1.2345e6789", token.Token{Type: token.FLOAT, Literal: "1.2345e6789", Start: 0, End: 11}},

		// Negative Floats with Exponent
		{"Negative float with lowercase exponent", "-1.2345e3", token.Token{Type: token.FLOAT, Literal: "-1.2345e3", Start: 0, End: 9}},
		{"Negative float with uppercase exponent", "-1.2345E3", token.Token{Type: token.FLOAT, Literal: "-1.2345E3", Start: 0, End: 9}},
		{"Negative float with positive exponent sign", "-1.2345e+3", token.Token{Type: token.FLOAT, Literal: "-1.2345e+3", Start: 0, End: 10}},
		{"Negative float with negative exponent sign", "-1.2345e-3", token.Token{Type: token.FLOAT, Literal: "-1.2345e-3", Start: 0, End: 10}},
		{"Negative float with zero exponent", "-1.2345e0", token.Token{Type: token.FLOAT, Literal: "-1.2345e0", Start: 0, End: 9}},
		{"Negative float with large exponent", "-1.2345e6789", token.Token{Type: token.FLOAT, Literal: "-1.2345e6789", Start: 0, End: 12}},
	})
}

func TestNextToken_IntNumbersFromZeroToTen(t *testing.T) {
	for i := 0; i <= 10; i++ {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			input := fmt.Sprintf("%d", i)
			l := New(input)
			tok, err := l.NextToken()

			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", input, err)
			}

			if tok.Type != token.INT {
				t.Errorf("expected token type INT, got %v", tok.Type)
			}

			if tok.Literal != input {
				t.Errorf("expected literal %q, got %q", input, tok.Literal)
			}
		})
	}
}

func TestNextToken_FloatNumbersFromZeroToTen(t *testing.T) {
	for i := 0; i <= 10; i++ {
		t.Run(fmt.Sprintf("%d.0", i), func(t *testing.T) {
			input := fmt.Sprintf("%d.0", i)
			l := New(input)
			tok, err := l.NextToken()

			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", input, err)
			}

			if tok.Type != token.FLOAT {
				t.Errorf("expected token type INT, got %v", tok.Type)
			}

			if tok.Literal != input {
				t.Errorf("expected literal %q, got %q", input, tok.Literal)
			}
		})
	}
}

func TestNextToken_InvalidNumbers(t *testing.T) {
	runInvalidTests(t, []struct {
		name        string
		input       string
		expectedErr error
	}{
		{"Invalid leading zero with extra digit", "00", &LexError{Line: 1, Column: 2, Err: errors.New("invalid number, unexpected digit after 0: '0'")}},
		{"Unfinished decimal after zero", "0.", &LexError{Line: 1, Column: 3, Err: errors.New("invalid number, expected digit but got '<EOF>'")}},
		{"Unexpected character at start", "*123", &LexError{Line: 1, Column: 1, Err: errors.New("unexpected character '*'")}},
		{"Plus sign as unexpected leading character", "+3", &LexError{Line: 1, Column: 1, Err: errors.New("unexpected character '+'")}},
		{"Double minus signs at start", "--123", &LexError{Line: 1, Column: 2, Err: errors.New("invalid number, expected digit but got '-'")}},
		{"Unexpected character after minus sign", "-*", &LexError{Line: 1, Column: 2, Err: errors.New("invalid number, expected digit but got '*'")}},
		{"Unexpected character within integer", "12x45", &LexError{Line: 1, Column: 3, Err: errors.New("invalid number, expected digit but got 'x'")}},
		{"Unexpected character after dot", "3.x", &LexError{Line: 1, Column: 3, Err: errors.New("invalid number, expected digit but got 'x'")}},
		{"Unexpected character after float", "3.1415x", &LexError{Line: 1, Column: 7, Err: errors.New("invalid number, expected digit but got 'x'")}},
		{"Unexpected character after minus sign", "-x", &LexError{Line: 1, Column: 2, Err: errors.New("invalid number, expected digit but got 'x'")}},
		{"Unexpected underscore in integer", "1_2345", &LexError{Line: 1, Column: 2, Err: errors.New("invalid number, expected digit but got '_'")}},
		{"Unexpected underscore in float", `3.14_15`, &LexError{Line: 1, Column: 5, Err: errors.New("invalid number, expected digit but got '_'")}},
		{"Unexpected character in exponent", "1.2345ex", &LexError{Line: 1, Column: 8, Err: errors.New("invalid number, expected digit but got 'x'")}},
		{"Unexpected character in exponent digits", "1.2e3x", &LexError{Line: 1, Column: 6, Err: errors.New("invalid number, expected digit but got 'x'")}},
		{"Unexpected quote in exponent", `1.2345e"`, &LexError{Line: 1, Column: 8, Err: errors.New(`invalid number, expected digit but got '\"'`)}},
		{"Unexpected dot in exponent", "1.2e3.", &LexError{Line: 1, Column: 6, Err: errors.New("invalid number, expected digit but got '.'")}},

		// Incorrect use of dot
		{"Incomplete float with trailing dot", "3.", &LexError{Line: 1, Column: 3, Err: errors.New("invalid number, expected digit but got '<EOF>'")}},
		{"Missing integer part", ".1415", &LexError{Line: 1, Column: 1, Err: errors.New("invalid number, expected digit before '.'")}},
		{"Double dots", "3..14", &LexError{Line: 1, Column: 3, Err: errors.New("invalid number, expected digit but got '.'")}},
		{"Multiple dots", "3.1.4", &LexError{Line: 1, Column: 4, Err: errors.New("invalid number, expected digit but got '.'")}},
		{"Trailing dot after float", "3.1415.", &LexError{Line: 1, Column: 7, Err: errors.New("invalid number, expected digit but got '.'")}},

		// Incomplete exponents
		{"No digits after lowercase exponent", "12345e", &LexError{Line: 1, Column: 7, Err: errors.New("invalid number, expected digit but got '<EOF>'")}},
		{"No digits after uppercase exponent", "12345E", &LexError{Line: 1, Column: 7, Err: errors.New("invalid number, expected digit but got '<EOF>'")}},
		{"No digits after exponent with plus sign", "1.2e+", &LexError{Line: 1, Column: 6, Err: errors.New("invalid number, expected digit but got '<EOF>'")}},
		{"No digits after exponent with minus sign", "1.2e-", &LexError{Line: 1, Column: 6, Err: errors.New("invalid number, expected digit but got '<EOF>'")}},
		{"Whitespace after exponent with plus sign ", "1.2e+ ", &LexError{Line: 1, Column: 6, Err: errors.New("invalid number, expected digit but got ' '")}},
		{"Whitespace after exponent with minus sign", "1.2e- ", &LexError{Line: 1, Column: 6, Err: errors.New("invalid number, expected digit but got ' '")}},

		// Incorrect exponent formats
		{"Double lowercase exponent", "1.2345e3e", &LexError{Line: 1, Column: 9, Err: errors.New("invalid number, expected digit but got 'e'")}},
		{"Mixed exponents (e and E)", "1.2345e3E", &LexError{Line: 1, Column: 9, Err: errors.New("invalid number, expected digit but got 'E'")}},
		{"Mixed exponents (E and e)", "1.2345E3e", &LexError{Line: 1, Column: 9, Err: errors.New("invalid number, expected digit but got 'e'")}},
		{"Double uppercase exponent", "1.2345E3E", &LexError{Line: 1, Column: 9, Err: errors.New("invalid number, expected digit but got 'E'")}},
		{"Immediate exponent after dot", "12345.e3", &LexError{Line: 1, Column: 7, Err: errors.New("invalid number, expected digit but got 'e'")}},
		{"Immediate uppercase exponent after dot", "12345.E3", &LexError{Line: 1, Column: 7, Err: errors.New("invalid number, expected digit but got 'E'")}},

		// Incorrect signs in exponent
		{"Double plus signs in exponent", "1.2e++3", &LexError{Line: 1, Column: 6, Err: errors.New("invalid number, expected digit but got '+'")}},
		{"Double minus signs in exponent", "1.2e--3", &LexError{Line: 1, Column: 6, Err: errors.New("invalid number, expected digit but got '-'")}},
		{"Mixed signs in exponent (+-)", "1.2e+-3", &LexError{Line: 1, Column: 6, Err: errors.New("invalid number, expected digit but got '-'")}},
		{"Mixed signs in exponent (-+)", "1.2e-+3", &LexError{Line: 1, Column: 6, Err: errors.New("invalid number, expected digit but got '+'")}},
	})
}

func TestNextToken_ValidStrings(t *testing.T) {
	runValidTests(t, []struct {
		name          string
		input         string
		expectedToken token.Token
	}{
		{"Empty", `""`, token.Token{Type: token.STRING, Literal: "", Start: 0, End: 2}},
		{"Simple", `"hello"`, token.Token{Type: token.STRING, Literal: "hello", Start: 0, End: 7}},
		{"Surrounded by whitespace", `" hello world "`, token.Token{Type: token.STRING, Literal: " hello world ", Start: 0, End: 15}},

		// Escaped Characters
		{"Escaped quote", `"hello \""`, token.Token{Type: token.STRING, Literal: `hello "`, Start: 0, End: 10}},
		{"Escaped slashes", `"hello \\ \\\\ \/"`, token.Token{Type: token.STRING, Literal: `hello \ \\ /`, Start: 0, End: 18}},
		{"Escaped control characters", `"hello \b\f\n\r\t"`, token.Token{Type: token.STRING, Literal: "hello \b\f\n\r\t", Start: 0, End: 18}},

		// Unicode in BMP
		{"Fixed-width escaped Unicode sequences", `"hello \u0123\u4567\u89AB\uCDEF"`, token.Token{Type: token.STRING, Literal: "hello \u0123\u4567\u89AB\uCDEF", Start: 0, End: 32}},
		{"Variable-width escaped Unicode sequences", `"hello \u{0123}\u{4567}\u{89AB}\u{CDEF}"`, token.Token{Type: token.STRING, Literal: "hello \u0123\u4567\u89AB\uCDEF", Start: 0, End: 40}},
		{"Fixed-width escaped Unicode with minimum width", `"hello \u0000"`, token.Token{Type: token.STRING, Literal: "hello \u0000", Start: 0, End: 14}},
		{"Variable-width escaped Unicode with minimum width", `"hello \u{0}"`, token.Token{Type: token.STRING, Literal: "hello \u0000", Start: 0, End: 13}},
		{"Zero-padded escaped Unicode with full width", `"hello \u{00000000}"`, token.Token{Type: token.STRING, Literal: "hello \u0000", Start: 0, End: 20}},

		// Unicode beyond BMP
		{"Unescaped Unicode beyond BMP", `"hello 🫶"`, token.Token{Type: token.STRING, Literal: "hello 🫶", Start: 0, End: 12}},
		{"Escaped Unicode beyond BMP", `"hello \u{1F60E}"`, token.Token{Type: token.STRING, Literal: "hello \U0001F60E", Start: 0, End: 17}},
		{"Maximum unescaped Unicode beyond BMP", "\"hello \U0010FFFF\"", token.Token{Type: token.STRING, Literal: "hello \U0010FFFF", Start: 0, End: 12}},
		{"Maximum escaped Unicode", `"hello \u{10FFFF}"`, token.Token{Type: token.STRING, Literal: "hello \U0010FFFF", Start: 0, End: 18}},

		// Surrogate Pairs
		{"Surrogate pair (heart emoji)", `"hello \uD83C\uDF0D"`, token.Token{Type: token.STRING, Literal: "hello 🌍", Start: 0, End: 20}},
		{"Minimum surrogate pair", `"hello \uD800\uDC00"`, token.Token{Type: token.STRING, Literal: "hello \U00010000", Start: 0, End: 20}},
		{"Maximum surrogate pair", `"hello \uDBFF\uDFFF"`, token.Token{Type: token.STRING, Literal: "hello \U0010FFFF", Start: 0, End: 20}},
	})
}

func TestNextToken_InvalidStrings(t *testing.T) {
	runInvalidTests(t, []struct {
		name        string
		input       string
		expectedErr error
	}{
		// Unterminated Strings
		{"Unterminated", `"`, &LexError{Line: 1, Column: 2, Err: errors.New("unterminated string")}},
		{"Unterminated with missing closing quote", `"hello world`, &LexError{Line: 1, Column: 13, Err: errors.New("unterminated string")}},
		{"Unterminated with newline", "\"hello\nworld\"", &LexError{Line: 1, Column: 7, Err: errors.New(`unterminated string`)}},
		{"Unterminated with carriage return", "\"hello\rworld\"", &LexError{Line: 1, Column: 7, Err: errors.New(`unterminated string`)}},

		{"Invalid quote character", `'hello world'`, &LexError{Line: 1, Column: 1, Err: errors.New("unexpected character '''")}},

		// Invalid Unicode Escape Sequences
		{"Unknown escape character", `"hello \x"`, &LexError{Line: 1, Column: 8, Err: errors.New(`unknown escape sequence '\x'`)}},
		{"Incomplete Unicode escape sequence", `"hello \u1 unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`invalid hex digit ' ' in Unicode escape sequence '\u1 '`)}},
		{"Invalid character in fixed-width Unicode escape", `"hello \u1Y34 unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`invalid hex digit 'Y' in Unicode escape sequence '\u1Y'`)}},
		{"Empty Unicode escape sequence", `"hello \u{} unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`unicode escape sequence cannot be empty`)}},
		{"Invalid character in variable-width Unicode escape", `"hello \u{1Y34} unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`invalid hex digit 'Y' in Unicode escape sequence '\u{1Y'`)}},
		{"Unclosed variable-width Unicode escape", `"hello \u{1234 unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`invalid hex digit ' ' in Unicode escape sequence '\u{1234 '`)}},
		{"Incomplete Unicode escape at EOF", `"hello \u{1234"`, &LexError{Line: 1, Column: 8, Err: errors.New(`invalid hex digit '"' in Unicode escape sequence '\u{1234"'`)}},

		// Out of Range or Invalid Unicode Escapes
		{"Invalid fixed-width Unicode escape", `"hello \uDEAD unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`invalid Unicode escape sequence '\uDEAD'`)}},
		{"Invalid variable-width Unicode escape", `"hello \u{DEAD} unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`unicode escape sequence '\u{DEAD}' is out of range or invalid`)}},
		{"Out-of-range Unicode escape", `"hello \u{110000} unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`unicode escape sequence '\u{110000}' is out of range or invalid`)}},
		{"Unicode escape value too high", `"hello \u{12345678} unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`unicode escape sequence '\u{12345678}' is out of range or invalid`)}},
		{"Unicode escape sequence too long", `"hello \u{000000000} unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`unicode escape sequence '\u{000000000' is too long`)}},

		// Invalid Surrogate Pairs
		{"Surrogates with braces not allowed", `"hello \u{D83D}\u{DE00} unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`unicode escape sequence '\u{D83D}' is out of range or invalid`)}},
		{"Invalid high surrogate pair", `"hello \uDEAD\uDEAD unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`invalid Unicode escape sequence '\uDEAD'`)}},
		{"Invalid low surrogate pair", `"hello \uD800\uD800 unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`invalid trailing surrogate in Unicode escape sequence '\uD800\uD800'`)}},
		{"Unexpected character after high surrogate", `"hello \uD83D\aDE00 unicode"`, &LexError{Line: 1, Column: 8, Err: errors.New(`expected 'u' after '\' in Unicode escape sequence`)}},
	})
}

func TestNextToken_ValidBlockStrings(t *testing.T) {
	runValidTests(t, []struct {
		name          string
		input         string
		expectedToken token.Token
	}{
		{"Empty", `""""""`, token.Token{Type: token.BLOCK_STRING, Literal: ``, Start: 0, End: 6}},
		{"Simple", `"""hello"""`, token.Token{Type: token.BLOCK_STRING, Literal: `hello`, Start: 0, End: 11}},
		{"Surrounded by whitespace", `""" hello world """`, token.Token{Type: token.BLOCK_STRING, Literal: ` hello world `, Start: 0, End: 19}},
		{"Quote", `"""hello " world"""`, token.Token{Type: token.BLOCK_STRING, Literal: `hello " world`, Start: 0, End: 19}},
		{"Triple quotes", `"""hello \""" world"""`, token.Token{Type: token.BLOCK_STRING, Literal: `hello """ world`, Start: 0, End: 22}},
		{"Newlines", "\"\"\"hello\nworld\"\"\"", token.Token{Type: token.BLOCK_STRING, Literal: "hello\nworld", Start: 0, End: 17}},
		{"Normalized newlines", "\"\"\"foo\rbar\r\nbaz\"\"\"", token.Token{Type: token.BLOCK_STRING, Literal: "foo\nbar\nbaz", Start: 0, End: 18}},
		{"Slashes", `"""hello \ /"""`, token.Token{Type: token.BLOCK_STRING, Literal: `hello \ /`, Start: 0, End: 15}},
		{"Unescaped control characters", `"""hello \b\f\n\r\t"""`, token.Token{Type: token.BLOCK_STRING, Literal: `hello \b\f\n\r\t`, Start: 0, End: 22}},
		{"Unescaped Unicode", `"""hello 🫶"""`, token.Token{Type: token.BLOCK_STRING, Literal: "hello 🫶", Start: 0, End: 16}},
		{"Multiple lines", `"""

        foo
            bar
                baz

        """`, token.Token{Type: token.BLOCK_STRING, Literal: "foo\n    bar\n        baz", Start: 0, End: 65}},
	})
}

func TestNextToken_InvalidBlockStrings(t *testing.T) {
	runInvalidTests(t, []struct {
		name        string
		input       string
		expectedErr error
	}{
		{"Unterminated", `"""`, &LexError{Line: 1, Column: 4, Err: errors.New("unterminated block string")}},
		{"Unterminated with missing closing quote", `"""hello world`, &LexError{Line: 1, Column: 15, Err: errors.New("unterminated block string")}},
	})

}

func TestNextToken_NextTokenAfterBlockStrings(t *testing.T) {
	input := `"""

            foo
                bar
                    baz

            """ next_token`

	expectedToken := token.Token{
		Type:    token.NAME,
		Literal: `next_token`,
		Start:   82,
		End:     92,
	}

	l := New(input)

	_, err := l.NextToken()
	if err != nil {
		t.Fatalf("unexpected error while reading block string: %v", err)
	}

	tok2, err := l.NextToken()
	if err != nil {
		t.Fatalf("unexpected error while reading second token: %v", err)
	}

	assertToken(t, tok2, expectedToken)
}

func TestNextToken_ValidComments(t *testing.T) {
	runValidTests(t, []struct {
		name          string
		input         string
		expectedToken token.Token
	}{
		{"Simple", "# hello", token.Token{Type: token.COMMENT, Literal: " hello", Start: 0, End: 7}},
		{"Newline", "# hello\nworld", token.Token{Type: token.COMMENT, Literal: " hello", Start: 0, End: 7}},
		{"Carriage return", "# hello\r\nworld", token.Token{Type: token.COMMENT, Literal: " hello", Start: 0, End: 7}},
		{"Emoji", "# hello 🫶", token.Token{Type: token.COMMENT, Literal: " hello 🫶", Start: 0, End: 12}},
	})
}

func TestNextToken_UnknownCharacters(t *testing.T) {
	runInvalidTests(t, []struct {
		name        string
		input       string
		expectedErr error
	}{
		{"Null", "\x00", &LexError{Line: 1, Column: 1, Err: errors.New("unexpected character 'U+0000'")}},
		{"Dot", `..`, &LexError{Line: 1, Column: 1, Err: errors.New("unexpected '.'")}},
		{"Tilde", `~`, &LexError{Line: 1, Column: 1, Err: errors.New("unexpected character '~'")}},
		{"Slash", `/`, &LexError{Line: 1, Column: 1, Err: errors.New(`unexpected character '/'`)}},
		{"Backslash", `\`, &LexError{Line: 1, Column: 1, Err: errors.New(`unexpected character '\'`)}},
		{"Backspace", "\b", &LexError{Line: 1, Column: 1, Err: errors.New("unexpected character 'U+0008'")}},
		{"Form Feed", "\f", &LexError{Line: 1, Column: 1, Err: errors.New("unexpected character 'U+000C'")}},
		{"Unicode", "\u00AA", &LexError{Line: 1, Column: 1, Err: errors.New(`unexpected character 'U+00AA'`)}},
		{"Emoji", "\U0001f60E", &LexError{Line: 1, Column: 1, Err: errors.New(`unexpected character 'U+1F60E'`)}},
	})
}

func TestNextToken_Query(t *testing.T) {
	input := `query {
           user(id: 123) {
               name
               friends(first: 10) {
                   edges {
                       node {
                           name
                       }
                   }
               }
           }
       }`

	expectedTokens := []token.Token{
		{Type: token.NAME, Literal: "query", Start: 0, End: 5},
		{Type: token.LBRACE, Start: 6, End: 7},
		{Type: token.NAME, Literal: "user", Start: 19, End: 23},
		{Type: token.LPAREN, Start: 23, End: 24},
		{Type: token.NAME, Literal: "id", Start: 24, End: 26},
		{Type: token.COLON, Start: 26, End: 27},
		{Type: token.INT, Literal: "123", Start: 28, End: 31},
		{Type: token.RPAREN, Start: 31, End: 32},
		{Type: token.LBRACE, Start: 33, End: 34},
		{Type: token.NAME, Literal: "name", Start: 50, End: 54},
		{Type: token.NAME, Literal: "friends", Start: 70, End: 77},
		{Type: token.LPAREN, Start: 77, End: 78},
		{Type: token.NAME, Literal: "first", Start: 78, End: 83},
		{Type: token.COLON, Start: 83, End: 84},
		{Type: token.INT, Literal: "10", Start: 85, End: 87},
		{Type: token.RPAREN, Start: 87, End: 88},
		{Type: token.LBRACE, Start: 89, End: 90},
		{Type: token.NAME, Literal: "edges", Start: 110, End: 115},
		{Type: token.LBRACE, Start: 116, End: 117},
		{Type: token.NAME, Literal: "node", Start: 141, End: 145},
		{Type: token.LBRACE, Start: 146, End: 147},
		{Type: token.NAME, Literal: "name", Start: 175, End: 179},
		{Type: token.RBRACE, Start: 203, End: 204},
		{Type: token.RBRACE, Start: 224, End: 225},
		{Type: token.RBRACE, Start: 241, End: 242},
		{Type: token.RBRACE, Start: 254, End: 255},
		{Type: token.RBRACE, Start: 263, End: 264},
	}

	l := New(input)
	for i, expected := range expectedTokens {
		t.Run(fmt.Sprintf("Token%d", i), func(t *testing.T) {
			tok, err := l.NextToken()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertToken(t, tok, expected)
		})
	}
}

func assertToken(t *testing.T, actual, expected token.Token) {
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("Tokens do not match.\nWant: %+v\nGot:      %+v", expected, actual)
	}
}

func assertError(t *testing.T, err, expectedErr error) {
	if err == nil && expectedErr == nil {
		return
	}
	if err == nil {
		t.Fatalf("expected error %v, got nil", expectedErr)
	}
	if expectedErr == nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err.Error() != expectedErr.Error() {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}

func runValidTests(t *testing.T, tests []struct {
	name          string
	input         string
	expectedToken token.Token
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok, err := l.NextToken()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertToken(t, tok, tt.expectedToken)
		})
	}
}

func runInvalidTests(t *testing.T, tests []struct {
	name        string
	input       string
	expectedErr error
}) {
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			_, err := l.NextToken()
			assertError(t, err, tt.expectedErr)
		})
	}
}
