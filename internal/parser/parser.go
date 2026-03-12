// Package parser implements a hand-written recursive descent parser for WORNG.
// It consumes a []lexer.Token and produces a *ast.ProgramNode.
//
// The parser is tolerant: it never panics and always returns a (partial) AST
// even when syntax errors are encountered.
package parser

// TODO(Phase 1): implement parser
