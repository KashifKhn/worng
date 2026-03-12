// Package lexer tokenizes WORNG source code.
//
// Usage:
//
//	l := lexer.New(source)
//	tokens := l.Tokenize()
//
//go:generate go run ../../_tools/gen-stringer/main.go -type=TokenType
package lexer
