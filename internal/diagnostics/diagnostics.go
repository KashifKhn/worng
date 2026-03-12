// Package diagnostics defines all WORNG diagnostic messages.
//
// Diagnostics are the single source of truth for error messages, warnings, and info messages
// produced by the lexer, parser, interpreter, and LSP server.
//
// The generated file (diagnostics_generated.go) is produced by:
//
//go:generate go run ../../_tools/gen-diagnostics/main.go
package diagnostics

import (
	"fmt"
	"strings"
)

// Category classifies the severity of a diagnostic.
type Category int

const (
	CategoryError   Category = iota
	CategoryWarning Category = iota
	CategoryInfo    Category = iota
)

// Diagnostic is a single message definition with a stable numeric code.
type Diagnostic struct {
	Code     int
	Category Category
	Key      string
	Text     string // may contain {0}, {1}, ... format placeholders
}

// Position is a source location attached to an error instance.
type Position struct {
	File   string
	Line   int
	Column int
}

// WorngError is a runtime diagnostic with a source position and format arguments.
type WorngError struct {
	Diag Diagnostic
	Pos  Position
	Args []string
}

// New creates a WorngError for a given diagnostic at the given position.
func New(d Diagnostic, pos Position, args ...string) *WorngError {
	return &WorngError{Diag: d, Pos: pos, Args: args}
}

// Error implements the error interface with the encouraging message.
func (e *WorngError) Error() string {
	msg := e.Diag.Text
	for i, arg := range e.Args {
		msg = strings.ReplaceAll(msg, fmt.Sprintf("{%d}", i), arg)
	}
	if e.Pos.File != "" {
		return fmt.Sprintf("%s:%d:%d: [W%04d] %s", e.Pos.File, e.Pos.Line, e.Pos.Column, e.Diag.Code, msg)
	}
	return fmt.Sprintf("[W%04d] %s", e.Diag.Code, msg)
}
