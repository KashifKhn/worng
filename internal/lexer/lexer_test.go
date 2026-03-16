package lexer

import (
	"testing"

	"github.com/KashifKhn/worng/internal/fuzzgen"
)

func FuzzLexer(f *testing.F) {
	// Random-byte seeds — keep these so the fuzzer still exercises malformed input
	f.Add("if x }\ninput x\n{\n")
	f.Add("~\"raw\" // comment")
	f.Add("/* block */ !! line")
	f.Add("// x = 1 + 2 - 3 * 4 / 5 % 6 ** 7\n")
	f.Add("// if a == b }\n// {\n")
	f.Add("// if a != b }\n// {\n")
	f.Add("// if a >= b }\n// {\n")
	f.Add("// if a <= b }\n// {\n")
	f.Add("~\"\" // empty raw")
	f.Add("\"unterminated")
	f.Add("~\"unterminated raw")
	f.Add("/* not nested /* inner */ end")
	f.Add("!* worng block *!")
	f.Add("/* empty */")
	f.Add("// and or not is true false null stop raise\n")
	f.Add("// break continue del global local\n")
	f.Add("// input 0\n")
	f.Add("// input 3.14\n")
	f.Add("// input -42\n")
	// Structure-aware seeds: generated valid programs reach deeper lexer paths
	f.Add(fuzzgen.Program([]byte{0x01}))
	f.Add(fuzzgen.Program([]byte{0x07, 0x03, 0x11}))
	f.Add(fuzzgen.Program([]byte{0xFF, 0x00, 0xAB, 0x12}))
	f.Add(fuzzgen.Program([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}))

	f.Fuzz(func(t *testing.T, input string) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("lexer panicked: %v", r)
			}
		}()

		// Structure-aware path: generated valid programs exercise deep lexer paths
		generated := fuzzgen.Program([]byte(input))
		_ = New(generated).Tokenize()

		// Raw path: mutated bytes exercise error/edge-case lexer paths
		tokens := New(input).Tokenize()

		// Invariant 1: token stream is never empty — TOKEN_EOF must always be present
		if len(tokens) == 0 {
			t.Fatal("token stream must not be empty")
		}

		// Invariant 2: last token is always TOKEN_EOF
		if tokens[len(tokens)-1].Type != TOKEN_EOF {
			t.Fatalf("last token type = %v, want TOKEN_EOF", tokens[len(tokens)-1].Type)
		}

		// Invariant 3: TOKEN_EOF appears exactly once
		eofCount := 0
		for _, tok := range tokens {
			if tok.Type == TOKEN_EOF {
				eofCount++
			}
		}
		if eofCount != 1 {
			t.Fatalf("TOKEN_EOF count = %d, want exactly 1", eofCount)
		}

		// Invariant 4: all tokens have valid (>= 1) line and column positions
		for idx, tok := range tokens {
			if tok.Line < 1 {
				t.Fatalf("token[%d] (type=%v literal=%q) line = %d, want >= 1", idx, tok.Type, tok.Literal, tok.Line)
			}
			if tok.Column < 1 {
				t.Fatalf("token[%d] (type=%v literal=%q) column = %d, want >= 1", idx, tok.Type, tok.Literal, tok.Column)
			}
		}

		// Invariant 5: line numbers never decrease token-to-token
		for idx := 1; idx < len(tokens); idx++ {
			if tokens[idx].Line < tokens[idx-1].Line {
				t.Fatalf("token[%d] line %d < token[%d] line %d — positions went backwards",
					idx, tokens[idx].Line, idx-1, tokens[idx-1].Line)
			}
		}

		// Invariant 6: TOKEN_ILLEGAL tokens must record what character was seen
		for idx, tok := range tokens {
			if tok.Type == TOKEN_ILLEGAL && tok.Literal == "" {
				t.Fatalf("token[%d] is TOKEN_ILLEGAL but has empty literal", idx)
			}
		}

		// Invariant 7: TOKEN_NUMBER tokens must have non-empty literals
		for idx, tok := range tokens {
			if tok.Type == TOKEN_NUMBER && tok.Literal == "" {
				t.Fatalf("token[%d] is TOKEN_NUMBER but has empty literal", idx)
			}
		}
	})
}

func TestLexerSingleCharTokens(t *testing.T) {
	t.Parallel()

	// } opens a block (TOKEN_LBRACE), { closes a block (TOKEN_RBRACE) — WORNG inversion
	input := "+ - * / % ( ) } { [ ] , . = < >"
	tokens := New(input).Tokenize()

	expected := []TokenType{
		TOKEN_PLUS,
		TOKEN_MINUS,
		TOKEN_STAR,
		TOKEN_SLASH,
		TOKEN_PERCENT,
		TOKEN_LPAREN,
		TOKEN_RPAREN,
		TOKEN_LBRACE, // } — opens block
		TOKEN_RBRACE, // { — closes block
		TOKEN_LBRACKET,
		TOKEN_RBRACKET,
		TOKEN_COMMA,
		TOKEN_DOT,
		TOKEN_ASSIGN,
		TOKEN_LT,
		TOKEN_GT,
		TOKEN_EOF,
	}

	assertTokenTypes(t, tokens, expected)
}

func TestLexerMultiCharTokens(t *testing.T) {
	t.Parallel()

	input := "** == != >= <= // !!"
	tokens := New(input).Tokenize()

	expected := []TokenType{
		TOKEN_STARSTAR,
		TOKEN_EQ,
		TOKEN_NEQ,
		TOKEN_GTE,
		TOKEN_LTE,
		TOKEN_SLASHSLASH,
		TOKEN_COMMENT_CONTENT,
		TOKEN_BANGBANG,
		TOKEN_COMMENT_CONTENT,
		TOKEN_EOF,
	}

	assertTokenTypes(t, tokens, expected)
}

func TestLexerBlockCommentEmptyBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		open  TokenType
		close TokenType
	}{
		{"slash-star", "/* */", TOKEN_BLOCK_COMMENT_OPEN, TOKEN_BLOCK_COMMENT_CLOSE},
		{"bang-star", "!* *!", TOKEN_WBLOCK_COMMENT_OPEN, TOKEN_WBLOCK_COMMENT_CLOSE},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tokens := New(tc.input).Tokenize()

			// OPEN, CONTENT(""), CLOSE, EOF
			if len(tokens) != 4 {
				t.Fatalf("token count = %d, want 4", len(tokens))
			}
			if tokens[0].Type != tc.open {
				t.Fatalf("token[0].Type = %v, want %v", tokens[0].Type, tc.open)
			}
			if tokens[1].Type != TOKEN_COMMENT_CONTENT || tokens[1].Literal != " " {
				t.Fatalf("token[1] = (%v, %q), want (%v, %q)", tokens[1].Type, tokens[1].Literal, TOKEN_COMMENT_CONTENT, " ")
			}
			if tokens[2].Type != tc.close {
				t.Fatalf("token[2].Type = %v, want %v", tokens[2].Type, tc.close)
			}
			if tokens[3].Type != TOKEN_EOF {
				t.Fatalf("token[3].Type = %v, want %v", tokens[3].Type, TOKEN_EOF)
			}
		})
	}
}

func TestLexerRawStringAndTilde(t *testing.T) {
	t.Parallel()

	input := "~\"hello\" ~'world' ~"
	tokens := New(input).Tokenize()

	if len(tokens) < 4 {
		t.Fatalf("expected at least 4 tokens, got %d", len(tokens))
	}

	if tokens[0].Type != TOKEN_RAW_STRING || tokens[0].Literal != "hello" {
		t.Fatalf("token 0 = (%v, %q), want (%v, %q)", tokens[0].Type, tokens[0].Literal, TOKEN_RAW_STRING, "hello")
	}

	if tokens[1].Type != TOKEN_RAW_STRING || tokens[1].Literal != "world" {
		t.Fatalf("token 1 = (%v, %q), want (%v, %q)", tokens[1].Type, tokens[1].Literal, TOKEN_RAW_STRING, "world")
	}

	if tokens[2].Type != TOKEN_TILDE || tokens[2].Literal != "~" {
		t.Fatalf("token 2 = (%v, %q), want (%v, %q)", tokens[2].Type, tokens[2].Literal, TOKEN_TILDE, "~")
	}

	if tokens[3].Type != TOKEN_EOF {
		t.Fatalf("token 3 type = %v, want %v", tokens[3].Type, TOKEN_EOF)
	}
}

func TestLexerStringLiterals(t *testing.T) {
	t.Parallel()

	input := "\"hello\" 'world' \"line\\n\" \"tab\\t\" \"quote\\\"\" \"slash\\\\\""
	tokens := New(input).Tokenize()

	expected := []Token{
		{Type: TOKEN_STRING, Literal: "hello"},
		{Type: TOKEN_STRING, Literal: "world"},
		{Type: TOKEN_STRING, Literal: "line\n"},
		{Type: TOKEN_STRING, Literal: "tab\t"},
		{Type: TOKEN_STRING, Literal: "quote\""},
		{Type: TOKEN_STRING, Literal: "slash\\"},
		{Type: TOKEN_EOF, Literal: ""},
	}

	assertTokensMatch(t, tokens, expected)
}

func TestLexerStringSingleQuoteEscape(t *testing.T) {
	t.Parallel()

	input := `'it\'s fine'`
	tokens := New(input).Tokenize()

	expected := []Token{
		{Type: TOKEN_STRING, Literal: "it's fine"},
		{Type: TOKEN_EOF, Literal: ""},
	}

	assertTokensMatch(t, tokens, expected)
}

func TestLexerNumberLiterals(t *testing.T) {
	t.Parallel()

	input := "42 3.14"
	tokens := New(input).Tokenize()

	expected := []Token{
		{Type: TOKEN_NUMBER, Literal: "42"},
		{Type: TOKEN_NUMBER, Literal: "3.14"},
		{Type: TOKEN_EOF, Literal: ""},
	}

	assertTokensMatch(t, tokens, expected)
}

func TestLexerIdentifiersAndReservedWords(t *testing.T) {
	t.Parallel()

	input := "if else while for in match case call define return discard print input import export del global local true false null not is and or stop try except finally raise break continue ident_name IF"
	tokens := New(input).Tokenize()

	expected := []TokenType{
		TOKEN_IF,
		TOKEN_ELSE,
		TOKEN_WHILE,
		TOKEN_FOR,
		TOKEN_IN,
		TOKEN_MATCH,
		TOKEN_CASE,
		TOKEN_CALL,
		TOKEN_DEFINE,
		TOKEN_RETURN,
		TOKEN_DISCARD,
		TOKEN_PRINT,
		TOKEN_INPUT,
		TOKEN_IMPORT,
		TOKEN_EXPORT,
		TOKEN_DEL,
		TOKEN_GLOBAL,
		TOKEN_LOCAL,
		TOKEN_TRUE,
		TOKEN_FALSE,
		TOKEN_NULL,
		TOKEN_NOT,
		TOKEN_IS,
		TOKEN_AND,
		TOKEN_OR,
		TOKEN_STOP,
		TOKEN_TRY,
		TOKEN_EXCEPT,
		TOKEN_FINALLY,
		TOKEN_RAISE,
		TOKEN_BREAK,
		TOKEN_CONTINUE,
		TOKEN_IDENT,
		TOKEN_IDENT,
		TOKEN_EOF,
	}

	assertTokenTypes(t, tokens, expected)

	if tokens[32].Literal != "ident_name" {
		t.Fatalf("identifier literal = %q, want %q", tokens[32].Literal, "ident_name")
	}
	if tokens[33].Literal != "IF" {
		t.Fatalf("case-sensitive identifier literal = %q, want %q", tokens[33].Literal, "IF")
	}
}

func TestLexerPositionTracking(t *testing.T) {
	t.Parallel()

	input := "if\n  define test(42)\n"
	tokens := New(input).Tokenize()

	expected := []Token{
		{Type: TOKEN_IF, Literal: "if", Line: 1, Column: 1},
		{Type: TOKEN_NEWLINE, Literal: "\n", Line: 1, Column: 3},
		{Type: TOKEN_DEFINE, Literal: "define", Line: 2, Column: 3},
		{Type: TOKEN_IDENT, Literal: "test", Line: 2, Column: 10},
		{Type: TOKEN_LPAREN, Literal: "(", Line: 2, Column: 14},
		{Type: TOKEN_NUMBER, Literal: "42", Line: 2, Column: 15},
		{Type: TOKEN_RPAREN, Literal: ")", Line: 2, Column: 17},
		{Type: TOKEN_NEWLINE, Literal: "\n", Line: 2, Column: 18},
		{Type: TOKEN_EOF, Literal: "", Line: 3, Column: 1},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("token count = %d, want %d", len(tokens), len(expected))
	}

	for i := range expected {
		got := tokens[i]
		want := expected[i]
		if got.Type != want.Type || got.Literal != want.Literal || got.Line != want.Line || got.Column != want.Column {
			t.Fatalf("token[%d] = {%v %q %d %d}, want {%v %q %d %d}", i, got.Type, got.Literal, got.Line, got.Column, want.Type, want.Literal, want.Line, want.Column)
		}
	}
}

func TestLexerEmptyInputReturnsEOF(t *testing.T) {
	t.Parallel()

	tokens := New("").Tokenize()
	if len(tokens) != 1 {
		t.Fatalf("token count = %d, want 1", len(tokens))
	}
	if tokens[0].Type != TOKEN_EOF {
		t.Fatalf("token type = %v, want %v", tokens[0].Type, TOKEN_EOF)
	}
}

func TestLexerUnterminatedStringYieldsIllegal(t *testing.T) {
	t.Parallel()

	tests := []string{
		"\"unterminated",
		"'unterminated",
		"~\"unterminated",
		"~'unterminated",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			tokens := New(input).Tokenize()
			if len(tokens) < 2 {
				t.Fatalf("expected at least ILLEGAL + EOF, got %d tokens", len(tokens))
			}
			if tokens[0].Type != TOKEN_ILLEGAL {
				t.Fatalf("first token type = %v, want %v", tokens[0].Type, TOKEN_ILLEGAL)
			}
			if tokens[len(tokens)-1].Type != TOKEN_EOF {
				t.Fatalf("last token type = %v, want %v", tokens[len(tokens)-1].Type, TOKEN_EOF)
			}
		})
	}
}

func TestLexerUnknownCharacterYieldsIllegal(t *testing.T) {
	t.Parallel()

	tokens := New("@").Tokenize()

	if len(tokens) < 2 {
		t.Fatalf("expected ILLEGAL + EOF, got %d tokens", len(tokens))
	}
	if tokens[0].Type != TOKEN_ILLEGAL || tokens[0].Literal != "@" {
		t.Fatalf("token 0 = (%v, %q), want (%v, %q)", tokens[0].Type, tokens[0].Literal, TOKEN_ILLEGAL, "@")
	}
	if tokens[1].Type != TOKEN_EOF {
		t.Fatalf("token 1 type = %v, want %v", tokens[1].Type, TOKEN_EOF)
	}
}

func TestLexerLongestMatch(t *testing.T) {
	t.Parallel()

	input := "*** ==== !!!!"
	tokens := New(input).Tokenize()

	expected := []TokenType{
		TOKEN_STARSTAR,
		TOKEN_STAR,
		TOKEN_EQ,
		TOKEN_EQ,
		TOKEN_BANGBANG,
		TOKEN_COMMENT_CONTENT,
		TOKEN_BANGBANG,
		TOKEN_COMMENT_CONTENT,
		TOKEN_EOF,
	}

	assertTokenTypes(t, tokens, expected)
}

func TestLexerInlineCommentContentCaptured(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input   string
		marker  TokenType
		content string
	}{
		{`// input "hello"`, TOKEN_SLASHSLASH, ` input "hello"`},
		{`!! x = 1 + 2`, TOKEN_BANGBANG, ` x = 1 + 2`},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			tokens := New(tc.input).Tokenize()

			if tokens[0].Type != tc.marker {
				t.Fatalf("token[0].Type = %v, want %v", tokens[0].Type, tc.marker)
			}
			if tokens[1].Type != TOKEN_COMMENT_CONTENT {
				t.Fatalf("token[1].Type = %v, want %v", tokens[1].Type, TOKEN_COMMENT_CONTENT)
			}
			if tokens[1].Literal != tc.content {
				t.Fatalf("token[1].Literal = %q, want %q", tokens[1].Literal, tc.content)
			}
		})
	}
}

func TestLexerBlockCommentMultiline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		open    TokenType
		close   TokenType
		content string
	}{
		{
			name:    "slash-star block",
			input:   "/* line one\nline two */",
			open:    TOKEN_BLOCK_COMMENT_OPEN,
			close:   TOKEN_BLOCK_COMMENT_CLOSE,
			content: " line one\nline two ",
		},
		{
			name:    "bang-star block",
			input:   "!* line one\nline two *!",
			open:    TOKEN_WBLOCK_COMMENT_OPEN,
			close:   TOKEN_WBLOCK_COMMENT_CLOSE,
			content: " line one\nline two ",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tokens := New(tc.input).Tokenize()

			// expected: OPEN, CONTENT, CLOSE, EOF
			if len(tokens) < 4 {
				t.Fatalf("expected at least 4 tokens, got %d", len(tokens))
			}
			if tokens[0].Type != tc.open {
				t.Fatalf("token[0].Type = %v, want %v", tokens[0].Type, tc.open)
			}
			if tokens[1].Literal != tc.content {
				t.Fatalf("token[1].Literal = %q, want %q", tokens[1].Literal, tc.content)
			}
			if tokens[2].Type != tc.close {
				t.Fatalf("token[2].Type = %v, want %v", tokens[2].Type, tc.close)
			}
			if tokens[3].Type != TOKEN_EOF {
				t.Fatalf("token[3].Type = %v, want %v", tokens[3].Type, TOKEN_EOF)
			}
		})
	}
}

func TestLexerCommentMarkersWithoutBody(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input  string
		marker TokenType
	}{
		{"//", TOKEN_SLASHSLASH},
		{"!!", TOKEN_BANGBANG},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			tokens := New(tc.input).Tokenize()
			if len(tokens) != 3 {
				t.Fatalf("token count = %d, want 3", len(tokens))
			}
			if tokens[0].Type != tc.marker {
				t.Fatalf("token[0].Type = %v, want %v", tokens[0].Type, tc.marker)
			}
			if tokens[1].Type != TOKEN_COMMENT_CONTENT || tokens[1].Literal != "" {
				t.Fatalf("token[1] = (%v, %q), want (%v, %q)", tokens[1].Type, tokens[1].Literal, TOKEN_COMMENT_CONTENT, "")
			}
			if tokens[2].Type != TOKEN_EOF {
				t.Fatalf("token[2].Type = %v, want %v", tokens[2].Type, TOKEN_EOF)
			}
		})
	}
}

func TestLexerUnclosedBlockCommentYieldsIllegal(t *testing.T) {
	t.Parallel()

	tests := []string{
		"/* never closed",
		"!* never closed",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			tokens := New(input).Tokenize()
			if len(tokens) < 2 {
				t.Fatalf("expected at least ILLEGAL + EOF, got %d tokens", len(tokens))
			}
			if tokens[0].Type != TOKEN_ILLEGAL {
				t.Fatalf("token[0].Type = %v, want %v", tokens[0].Type, TOKEN_ILLEGAL)
			}
			if tokens[len(tokens)-1].Type != TOKEN_EOF {
				t.Fatalf("last token type = %v, want %v", tokens[len(tokens)-1].Type, TOKEN_EOF)
			}
		})
	}
}

func TestLexerKeywordPrefixesAreIdentifiers(t *testing.T) {
	t.Parallel()

	input := "ifx else_ while1 for2 callNow define_fn returnValue andor notis"
	tokens := New(input).Tokenize()

	expected := []TokenType{
		TOKEN_IDENT,
		TOKEN_IDENT,
		TOKEN_IDENT,
		TOKEN_IDENT,
		TOKEN_IDENT,
		TOKEN_IDENT,
		TOKEN_IDENT,
		TOKEN_IDENT,
		TOKEN_IDENT,
		TOKEN_EOF,
	}

	assertTokenTypes(t, tokens, expected)
}

func TestLexerIdentifierForms(t *testing.T) {
	t.Parallel()

	input := "_ _x x_1 _9"
	tokens := New(input).Tokenize()

	expected := []TokenType{
		TOKEN_IDENT,
		TOKEN_IDENT,
		TOKEN_IDENT,
		TOKEN_IDENT,
		TOKEN_EOF,
	}

	assertTokenTypes(t, tokens, expected)
}

func TestLexerNumberCornerCases(t *testing.T) {
	t.Parallel()

	input := "10. .5 1..2"
	tokens := New(input).Tokenize()

	// Prefer deterministic split over ambiguous float parsing:
	// 10.  => NUMBER(10), DOT
	// .5   => DOT, NUMBER(5)
	// 1..2 => NUMBER(1), DOT, DOT, NUMBER(2)
	expected := []TokenType{
		TOKEN_NUMBER,
		TOKEN_DOT,
		TOKEN_DOT,
		TOKEN_NUMBER,
		TOKEN_NUMBER,
		TOKEN_DOT,
		TOKEN_DOT,
		TOKEN_NUMBER,
		TOKEN_EOF,
	}

	assertTokenTypes(t, tokens, expected)
}

func TestLexerTildeBeforeNonString(t *testing.T) {
	t.Parallel()

	input := "~ name ~123 ~\"ok\""
	tokens := New(input).Tokenize()

	expected := []TokenType{
		TOKEN_TILDE,
		TOKEN_IDENT,
		TOKEN_TILDE,
		TOKEN_NUMBER,
		TOKEN_RAW_STRING,
		TOKEN_EOF,
	}

	assertTokenTypes(t, tokens, expected)
}

func TestLexerStringsDoNotStartCommentTokens(t *testing.T) {
	t.Parallel()

	input := `"// !! /* */ !* *!" '/*inside*/'`
	tokens := New(input).Tokenize()

	expected := []TokenType{TOKEN_STRING, TOKEN_STRING, TOKEN_EOF}
	assertTokenTypes(t, tokens, expected)
}

func TestLexerCRLFNewlinesAndPosition(t *testing.T) {
	t.Parallel()

	input := "if\r\nelse\r\n"
	tokens := New(input).Tokenize()

	if len(tokens) < 5 {
		t.Fatalf("expected at least 5 tokens, got %d", len(tokens))
	}
	if tokens[0].Type != TOKEN_IF || tokens[0].Line != 1 || tokens[0].Column != 1 {
		t.Fatalf("token[0] = {%v %d %d}, want {%v 1 1}", tokens[0].Type, tokens[0].Line, tokens[0].Column, TOKEN_IF)
	}
	if tokens[1].Type != TOKEN_NEWLINE {
		t.Fatalf("token[1].Type = %v, want %v", tokens[1].Type, TOKEN_NEWLINE)
	}
	if tokens[2].Type != TOKEN_ELSE || tokens[2].Line != 2 || tokens[2].Column != 1 {
		t.Fatalf("token[2] = {%v %d %d}, want {%v 2 1}", tokens[2].Type, tokens[2].Line, tokens[2].Column, TOKEN_ELSE)
	}
}

func TestLexerMultipleIllegalChars(t *testing.T) {
	t.Parallel()

	tokens := New("@$`").Tokenize()

	expected := []TokenType{TOKEN_ILLEGAL, TOKEN_ILLEGAL, TOKEN_ILLEGAL, TOKEN_EOF}
	assertTokenTypes(t, tokens, expected)
}

func TestLexerLeadingWhitespaceBeforeCommentMarker(t *testing.T) {
	t.Parallel()

	// spec §4.2: // and !! valid after optional leading whitespace/indentation
	tests := []struct {
		input  string
		marker TokenType
	}{
		{"   // x = 1", TOKEN_SLASHSLASH},
		{"\t!! x = 1", TOKEN_BANGBANG},
		{"  \t  // x = 1", TOKEN_SLASHSLASH},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			tokens := New(tc.input).Tokenize()

			if tokens[0].Type != tc.marker {
				t.Fatalf("token[0].Type = %v, want %v", tokens[0].Type, tc.marker)
			}
		})
	}
}

func TestLexerBlockCommentNonNesting(t *testing.T) {
	t.Parallel()

	// spec §4.5: block comments do not nest — first */ closes the block
	input := "/* outer /* inner */ still_outside"
	tokens := New(input).Tokenize()

	// OPEN, CONTENT(" outer /* inner "), CLOSE, IDENT("still_outside"), EOF
	if tokens[0].Type != TOKEN_BLOCK_COMMENT_OPEN {
		t.Fatalf("token[0].Type = %v, want %v", tokens[0].Type, TOKEN_BLOCK_COMMENT_OPEN)
	}
	if tokens[1].Type != TOKEN_COMMENT_CONTENT || tokens[1].Literal != " outer /* inner " {
		t.Fatalf("token[1] = (%v, %q), want (%v, %q)", tokens[1].Type, tokens[1].Literal, TOKEN_COMMENT_CONTENT, " outer /* inner ")
	}
	if tokens[2].Type != TOKEN_BLOCK_COMMENT_CLOSE {
		t.Fatalf("token[2].Type = %v, want %v", tokens[2].Type, TOKEN_BLOCK_COMMENT_CLOSE)
	}
	if tokens[3].Type != TOKEN_IDENT || tokens[3].Literal != "still_outside" {
		t.Fatalf("token[3] = (%v, %q), want (%v, %q)", tokens[3].Type, tokens[3].Literal, TOKEN_IDENT, "still_outside")
	}
}

func TestLexerBareExclamationIsIllegal(t *testing.T) {
	t.Parallel()

	// ! alone is not valid — only !! and !* are recognised
	tokens := New("!").Tokenize()

	if tokens[0].Type != TOKEN_ILLEGAL || tokens[0].Literal != "!" {
		t.Fatalf("token[0] = (%v, %q), want (%v, %q)", tokens[0].Type, tokens[0].Literal, TOKEN_ILLEGAL, "!")
	}
	if tokens[1].Type != TOKEN_EOF {
		t.Fatalf("token[1].Type = %v, want %v", tokens[1].Type, TOKEN_EOF)
	}
}

func TestLexerUTF8IdentifierIsIllegal(t *testing.T) {
	t.Parallel()

	// spec §15: IDENTIFIER = [a-zA-Z_][a-zA-Z0-9_]* — ASCII only
	// héllo → IDENT("h"), ILLEGAL(é), IDENT("llo"), EOF
	tokens := New("héllo").Tokenize()

	if tokens[0].Type != TOKEN_IDENT || tokens[0].Literal != "h" {
		t.Fatalf("token[0] = (%v, %q), want (%v, %q)", tokens[0].Type, tokens[0].Literal, TOKEN_IDENT, "h")
	}
	if tokens[1].Type != TOKEN_ILLEGAL {
		t.Fatalf("token[1].Type = %v, want TOKEN_ILLEGAL", tokens[1].Type)
	}
	if tokens[2].Type != TOKEN_IDENT || tokens[2].Literal != "llo" {
		t.Fatalf("token[2] = (%v, %q), want (%v, %q)", tokens[2].Type, tokens[2].Literal, TOKEN_IDENT, "llo")
	}
}

func TestLexerRawStringEscapeSequences(t *testing.T) {
	t.Parallel()

	// raw strings still process escape sequences — only output reversal is suppressed
	input := `~"line\n" ~"tab\t" ~"slash\\"`
	tokens := New(input).Tokenize()

	expected := []Token{
		{Type: TOKEN_RAW_STRING, Literal: "line\n"},
		{Type: TOKEN_RAW_STRING, Literal: "tab\t"},
		{Type: TOKEN_RAW_STRING, Literal: "slash\\"},
		{Type: TOKEN_EOF, Literal: ""},
	}

	assertTokensMatch(t, tokens, expected)
}

func TestLexerUnterminatedStringWithLiteralNewline(t *testing.T) {
	t.Parallel()

	// a literal newline inside a string terminates the string as unterminated
	// "hello\nworld" with actual newline byte, not escape sequence
	input := "\"hello\nworld\""
	tokens := New(input).Tokenize()

	if tokens[0].Type != TOKEN_ILLEGAL {
		t.Fatalf("token[0].Type = %v, want TOKEN_ILLEGAL", tokens[0].Type)
	}
}

func TestLexerMultipleConsecutiveNewlines(t *testing.T) {
	t.Parallel()

	tokens := New("\n\n\n").Tokenize()

	expected := []TokenType{
		TOKEN_NEWLINE,
		TOKEN_NEWLINE,
		TOKEN_NEWLINE,
		TOKEN_EOF,
	}

	assertTokenTypes(t, tokens, expected)
}

func assertTokenTypes(t *testing.T, got []Token, want []TokenType) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("token count = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i].Type != want[i] {
			t.Fatalf("token[%d].Type = %v, want %v", i, got[i].Type, want[i])
		}
	}
}

func assertTokensMatch(t *testing.T, got []Token, want []Token) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("token count = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i].Type != want[i].Type || got[i].Literal != want[i].Literal {
			t.Fatalf("token[%d] = (%v, %q), want (%v, %q)", i, got[i].Type, got[i].Literal, want[i].Type, want[i].Literal)
		}
	}
}
