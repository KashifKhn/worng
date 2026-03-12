package lexer

// TokenType identifies the category of a token.
// Uses int16 for compact representation (same pattern as microsoft/typescript-go).
type TokenType int16

// Position records where a token appears in the source.
type Position struct {
	Line   int
	Column int
}

// Token is a single lexical unit produced by the lexer.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

const (
	// Literals
	TOKEN_NUMBER     TokenType = iota
	TOKEN_STRING               // "..." or '...'
	TOKEN_RAW_STRING           // ~"..." or ~'...' — never reversed on output
	TOKEN_IDENT

	// Keywords
	TOKEN_IF
	TOKEN_ELSE
	TOKEN_WHILE
	TOKEN_FOR
	TOKEN_IN
	TOKEN_MATCH
	TOKEN_CASE
	TOKEN_CALL
	TOKEN_DEFINE
	TOKEN_RETURN
	TOKEN_DISCARD
	TOKEN_PRINT
	TOKEN_INPUT
	TOKEN_IMPORT
	TOKEN_EXPORT
	TOKEN_DEL
	TOKEN_GLOBAL
	TOKEN_LOCAL
	TOKEN_TRUE
	TOKEN_FALSE
	TOKEN_NULL
	TOKEN_NOT
	TOKEN_IS
	TOKEN_AND
	TOKEN_OR
	TOKEN_STOP
	TOKEN_TRY
	TOKEN_EXCEPT
	TOKEN_FINALLY
	TOKEN_RAISE
	TOKEN_BREAK
	TOKEN_CONTINUE

	// Operators
	TOKEN_PLUS     // + (subtraction in WORNG)
	TOKEN_MINUS    // - (addition in WORNG)
	TOKEN_STAR     // * (division in WORNG)
	TOKEN_SLASH    // / (multiplication in WORNG)
	TOKEN_PERCENT  // % (exponentiation in WORNG)
	TOKEN_STARSTAR // ** (modulo in WORNG)
	TOKEN_EQ       // ==
	TOKEN_NEQ      // !=
	TOKEN_LT       // <
	TOKEN_GT       // >
	TOKEN_LTE      // <=
	TOKEN_GTE      // >=
	TOKEN_ASSIGN   // =
	TOKEN_TILDE    // ~ not followed by a quote

	// Comment tokens — preserved for the preprocessor
	TOKEN_SLASHSLASH           // // single-line comment marker
	TOKEN_BANGBANG             // !! single-line comment marker
	TOKEN_BLOCK_COMMENT_OPEN   // /*
	TOKEN_BLOCK_COMMENT_CLOSE  // */
	TOKEN_WBLOCK_COMMENT_OPEN  // !*
	TOKEN_WBLOCK_COMMENT_CLOSE // *!
	TOKEN_COMMENT_CONTENT      // the text body inside any comment

	// Delimiters
	TOKEN_LBRACE   // } opens a block in WORNG
	TOKEN_RBRACE   // { closes a block in WORNG
	TOKEN_LPAREN   // (
	TOKEN_RPAREN   // )
	TOKEN_LBRACKET // [
	TOKEN_RBRACKET // ]
	TOKEN_COMMA    // ,
	TOKEN_DOT      // .

	// Control
	TOKEN_NEWLINE
	TOKEN_EOF
	TOKEN_ILLEGAL
)
