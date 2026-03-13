// Package lexer tokenizes WORNG source code.
//
// Usage:
//
//	l := lexer.New(source)
//	tokens := l.Tokenize()
package lexer

import (
	"strings"
	"unicode/utf8"
)

type Lexer struct {
	input   string
	pos     int
	line    int
	column  int
	pending []Token
	eofSent bool
}

func New(input string) *Lexer {
	return &Lexer{
		input:  input,
		line:   1,
		column: 1,
	}
}

func (l *Lexer) Tokenize() []Token {
	tokens := make([]Token, 0, len(l.input)/2+1)
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return tokens
}

func (l *Lexer) NextToken() Token {
	if len(l.pending) > 0 {
		tok := l.pending[0]
		l.pending = l.pending[1:]
		return tok
	}

	l.skipSpacesAndTabs()

	if l.atEOF() {
		if l.eofSent {
			return Token{Type: TOKEN_EOF, Line: l.line, Column: l.column}
		}
		l.eofSent = true
		return Token{Type: TOKEN_EOF, Line: l.line, Column: l.column}
	}

	startLine, startCol := l.line, l.column

	if l.matchString("\r\n") {
		l.consumeString("\r\n")
		return Token{Type: TOKEN_NEWLINE, Literal: "\n", Line: startLine, Column: startCol}
	}
	if l.matchByte('\n') {
		l.consumeRune()
		return Token{Type: TOKEN_NEWLINE, Literal: "\n", Line: startLine, Column: startCol}
	}

	if l.matchString("//") {
		return l.emitLineComment(TOKEN_SLASHSLASH, "//")
	}
	if l.matchString("!!") {
		return l.emitLineComment(TOKEN_BANGBANG, "!!")
	}

	if l.matchString("/*") {
		ok, content, close, okClose := l.readBlockComment("/*", "*/", TOKEN_BLOCK_COMMENT_OPEN, TOKEN_BLOCK_COMMENT_CLOSE)
		if !okClose {
			return ok
		}
		l.pending = append(l.pending, content, close)
		return ok
	}
	if l.matchString("!*") {
		ok, content, close, okClose := l.readBlockComment("!*", "*!", TOKEN_WBLOCK_COMMENT_OPEN, TOKEN_WBLOCK_COMMENT_CLOSE)
		if !okClose {
			return ok
		}
		l.pending = append(l.pending, content, close)
		return ok
	}

	if l.matchString("**") {
		l.consumeString("**")
		return Token{Type: TOKEN_STARSTAR, Literal: "**", Line: startLine, Column: startCol}
	}
	if l.matchString("==") {
		l.consumeString("==")
		return Token{Type: TOKEN_EQ, Literal: "==", Line: startLine, Column: startCol}
	}
	if l.matchString("!=") {
		l.consumeString("!=")
		return Token{Type: TOKEN_NEQ, Literal: "!=", Line: startLine, Column: startCol}
	}
	if l.matchString("<=") {
		l.consumeString("<=")
		return Token{Type: TOKEN_LTE, Literal: "<=", Line: startLine, Column: startCol}
	}
	if l.matchString(">=") {
		l.consumeString(">=")
		return Token{Type: TOKEN_GTE, Literal: ">=", Line: startLine, Column: startCol}
	}

	if l.matchByte('~') {
		l.consumeRune()
		if q, ok := l.peekRune(); ok && (q == '"' || q == '\'') {
			strTok, okString := l.readString(TOKEN_RAW_STRING)
			if !okString {
				return Token{Type: TOKEN_ILLEGAL, Literal: "~", Line: startLine, Column: startCol}
			}
			strTok.Line = startLine
			strTok.Column = startCol
			return strTok
		}
		return Token{Type: TOKEN_TILDE, Literal: "~", Line: startLine, Column: startCol}
	}

	if r, ok := l.peekRune(); ok {
		if isASCIILetter(r) || r == '_' {
			lit := l.readIdentifier()
			return Token{Type: lookupIdent(lit), Literal: lit, Line: startLine, Column: startCol}
		}
		if isASCIIDigit(r) {
			lit := l.readNumber()
			return Token{Type: TOKEN_NUMBER, Literal: lit, Line: startLine, Column: startCol}
		}
		if r == '"' || r == '\'' {
			tok, okString := l.readString(TOKEN_STRING)
			if !okString {
				return Token{Type: TOKEN_ILLEGAL, Literal: string(r), Line: startLine, Column: startCol}
			}
			tok.Line = startLine
			tok.Column = startCol
			return tok
		}
	}

	switch r, _ := l.peekRune(); r {
	case '+':
		l.consumeRune()
		return Token{Type: TOKEN_PLUS, Literal: "+", Line: startLine, Column: startCol}
	case '-':
		l.consumeRune()
		return Token{Type: TOKEN_MINUS, Literal: "-", Line: startLine, Column: startCol}
	case '*':
		l.consumeRune()
		return Token{Type: TOKEN_STAR, Literal: "*", Line: startLine, Column: startCol}
	case '/':
		l.consumeRune()
		return Token{Type: TOKEN_SLASH, Literal: "/", Line: startLine, Column: startCol}
	case '%':
		l.consumeRune()
		return Token{Type: TOKEN_PERCENT, Literal: "%", Line: startLine, Column: startCol}
	case '<':
		l.consumeRune()
		return Token{Type: TOKEN_LT, Literal: "<", Line: startLine, Column: startCol}
	case '>':
		l.consumeRune()
		return Token{Type: TOKEN_GT, Literal: ">", Line: startLine, Column: startCol}
	case '=':
		l.consumeRune()
		return Token{Type: TOKEN_ASSIGN, Literal: "=", Line: startLine, Column: startCol}
	case '}':
		l.consumeRune()
		return Token{Type: TOKEN_LBRACE, Literal: "}", Line: startLine, Column: startCol}
	case '{':
		l.consumeRune()
		return Token{Type: TOKEN_RBRACE, Literal: "{", Line: startLine, Column: startCol}
	case '(':
		l.consumeRune()
		return Token{Type: TOKEN_LPAREN, Literal: "(", Line: startLine, Column: startCol}
	case ')':
		l.consumeRune()
		return Token{Type: TOKEN_RPAREN, Literal: ")", Line: startLine, Column: startCol}
	case '[':
		l.consumeRune()
		return Token{Type: TOKEN_LBRACKET, Literal: "[", Line: startLine, Column: startCol}
	case ']':
		l.consumeRune()
		return Token{Type: TOKEN_RBRACKET, Literal: "]", Line: startLine, Column: startCol}
	case ',':
		l.consumeRune()
		return Token{Type: TOKEN_COMMA, Literal: ",", Line: startLine, Column: startCol}
	case '.':
		l.consumeRune()
		return Token{Type: TOKEN_DOT, Literal: ".", Line: startLine, Column: startCol}
	case '!':
		l.consumeRune()
		return Token{Type: TOKEN_ILLEGAL, Literal: "!", Line: startLine, Column: startCol}
	default:
		ch, _ := l.consumeRune()
		return Token{Type: TOKEN_ILLEGAL, Literal: string(ch), Line: startLine, Column: startCol}
	}
}

func (l *Lexer) emitLineComment(marker TokenType, markerLit string) Token {
	startLine, startCol := l.line, l.column
	l.consumeString(markerLit)

	contentLine, contentCol := l.line, l.column
	start := l.pos
	for !l.atEOF() {
		if l.matchString("\r\n") || l.matchByte('\n') {
			break
		}
		if l.matchString("//") || l.matchString("!!") {
			break
		}
		l.consumeRune()
	}
	content := l.input[start:l.pos]

	l.pending = append(l.pending, Token{Type: TOKEN_COMMENT_CONTENT, Literal: content, Line: contentLine, Column: contentCol})
	return Token{Type: marker, Literal: markerLit, Line: startLine, Column: startCol}
}

func (l *Lexer) readBlockComment(openLit, closeLit string, openType, closeType TokenType) (Token, Token, Token, bool) {
	openLine, openCol := l.line, l.column
	l.consumeString(openLit)

	contentLine, contentCol := l.line, l.column
	start := l.pos
	for !l.atEOF() {
		if l.matchString(closeLit) {
			content := l.input[start:l.pos]
			closeLine, closeCol := l.line, l.column
			l.consumeString(closeLit)
			return Token{Type: openType, Literal: openLit, Line: openLine, Column: openCol},
				Token{Type: TOKEN_COMMENT_CONTENT, Literal: content, Line: contentLine, Column: contentCol},
				Token{Type: closeType, Literal: closeLit, Line: closeLine, Column: closeCol},
				true
		}
		l.consumeRune()
	}

	return Token{Type: TOKEN_ILLEGAL, Literal: openLit, Line: openLine, Column: openCol}, Token{}, Token{}, false
}

func (l *Lexer) readIdentifier() string {
	start := l.pos
	for {
		r, ok := l.peekRune()
		if !ok || (!isASCIILetter(r) && !isASCIIDigit(r) && r != '_') {
			break
		}
		l.consumeRune()
	}
	return l.input[start:l.pos]
}

func (l *Lexer) readNumber() string {
	start := l.pos
	for {
		r, ok := l.peekRune()
		if !ok || !isASCIIDigit(r) {
			break
		}
		l.consumeRune()
	}

	if l.matchByte('.') {
		next, ok := l.peekRuneAt(l.pos + 1)
		if ok && isASCIIDigit(next) {
			l.consumeRune() // dot
			for {
				r, ok := l.peekRune()
				if !ok || !isASCIIDigit(r) {
					break
				}
				l.consumeRune()
			}
		}
	}

	return l.input[start:l.pos]
}

func (l *Lexer) readString(tokenType TokenType) (Token, bool) {
	q, ok := l.peekRune()
	if !ok || (q != '"' && q != '\'') {
		return Token{}, false
	}

	l.consumeRune() // opening quote
	var b strings.Builder

	for !l.atEOF() {
		r, _ := l.peekRune()
		if r == '\n' || r == '\r' {
			return Token{}, false
		}
		if r == q {
			l.consumeRune()
			return Token{Type: tokenType, Literal: b.String()}, true
		}
		if r == '\\' {
			l.consumeRune()
			esc, ok := l.peekRune()
			if !ok {
				return Token{}, false
			}
			switch esc {
			case 'n':
				b.WriteRune('\n')
			case 't':
				b.WriteRune('\t')
			case '\\':
				b.WriteRune('\\')
			case '"':
				b.WriteRune('"')
			case '\'':
				b.WriteRune('\'')
			default:
				b.WriteRune(esc)
			}
			l.consumeRune()
			continue
		}
		l.consumeRune()
		b.WriteRune(r)
	}

	return Token{}, false
}

func (l *Lexer) skipSpacesAndTabs() {
	for !l.atEOF() {
		r, ok := l.peekRune()
		if !ok || (r != ' ' && r != '\t') {
			return
		}
		l.consumeRune()
	}
}

func (l *Lexer) atEOF() bool {
	return l.pos >= len(l.input)
}

func (l *Lexer) matchByte(b byte) bool {
	return l.pos < len(l.input) && l.input[l.pos] == b
}

func (l *Lexer) matchString(s string) bool {
	return strings.HasPrefix(l.input[l.pos:], s)
}

func (l *Lexer) consumeString(s string) {
	for range s {
		l.consumeRune()
	}
}

func (l *Lexer) peekRune() (rune, bool) {
	if l.atEOF() {
		return 0, false
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])
	return r, true
}

func (l *Lexer) peekRuneAt(i int) (rune, bool) {
	if i >= len(l.input) {
		return 0, false
	}
	r, _ := utf8.DecodeRuneInString(l.input[i:])
	return r, true
}

func (l *Lexer) consumeRune() (rune, int) {
	if l.atEOF() {
		return 0, 0
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += w
	if r == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return r, w
}

func isASCIILetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

var keywords = map[string]TokenType{
	"if":       TOKEN_IF,
	"else":     TOKEN_ELSE,
	"while":    TOKEN_WHILE,
	"for":      TOKEN_FOR,
	"in":       TOKEN_IN,
	"match":    TOKEN_MATCH,
	"case":     TOKEN_CASE,
	"call":     TOKEN_CALL,
	"define":   TOKEN_DEFINE,
	"return":   TOKEN_RETURN,
	"discard":  TOKEN_DISCARD,
	"print":    TOKEN_PRINT,
	"input":    TOKEN_INPUT,
	"import":   TOKEN_IMPORT,
	"export":   TOKEN_EXPORT,
	"del":      TOKEN_DEL,
	"global":   TOKEN_GLOBAL,
	"local":    TOKEN_LOCAL,
	"true":     TOKEN_TRUE,
	"false":    TOKEN_FALSE,
	"null":     TOKEN_NULL,
	"not":      TOKEN_NOT,
	"is":       TOKEN_IS,
	"and":      TOKEN_AND,
	"or":       TOKEN_OR,
	"stop":     TOKEN_STOP,
	"try":      TOKEN_TRY,
	"except":   TOKEN_EXCEPT,
	"finally":  TOKEN_FINALLY,
	"raise":    TOKEN_RAISE,
	"break":    TOKEN_BREAK,
	"continue": TOKEN_CONTINUE,
}

func lookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENT
}
