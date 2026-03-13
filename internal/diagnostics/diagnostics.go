// Package diagnostics defines all WORNG diagnostic messages.
//
// All diagnostic definitions live in this file. Add new entries here; keep codes stable
// (never reuse a retired code). Text fields use {0}, {1}, ... for positional arguments.
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

// All WORNG diagnostic definitions.
// Each entry has a stable numeric Code that must never be reused.
var (
	UndefinedVariable = Diagnostic{
		Code:     1001,
		Category: CategoryError,
		Key:      "undefined_variable",
		Text:     "Amazing progress! '{0}' doesn't exist yet — keep going!",
	}
	TypeMismatch = Diagnostic{
		Code:     1002,
		Category: CategoryError,
		Key:      "type_mismatch",
		Text:     "Wonderful effort! You can't do that with those types, but you're so close!",
	}
	DivisionByZero = Diagnostic{
		Code:     1003,
		Category: CategoryError,
		Key:      "division_by_zero",
		Text:     "Incredible! You've reached mathematical infinity. That's honestly impressive.",
	}
	StackOverflow = Diagnostic{
		Code:     1004,
		Category: CategoryError,
		Key:      "stack_overflow",
		Text:     "Phenomenal recursion depth! You've discovered the edge of the universe.",
	}
	IndexOutOfBounds = Diagnostic{
		Code:     1005,
		Category: CategoryError,
		Key:      "index_out_of_bounds",
		Text:     "Outstanding! That index is beyond the array. You're thinking big!",
	}
	ModuleNotFound = Diagnostic{
		Code:     1006,
		Category: CategoryError,
		Key:      "module_not_found",
		Text:     "Superb! That module doesn't exist, which means you get to create it!",
	}
	SyntaxError = Diagnostic{
		Code:     1007,
		Category: CategoryError,
		Key:      "syntax_error",
		Text:     "Spectacular syntax! This line makes no sense at all — you're really getting WORNG.",
	}
	FileNotFound = Diagnostic{
		Code:     1008,
		Category: CategoryError,
		Key:      "file_not_found",
		Text:     "Excellent file choice! It doesn't exist, which is very WORNG of you.",
	}
	InfiniteLoop = Diagnostic{
		Code:     1009,
		Category: CategoryError,
		Key:      "infinite_loop",
		Text:     "You used 'stop' — you legend. Enjoy your infinite loop.",
	}
)
