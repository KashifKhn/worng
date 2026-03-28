// Package diagnostics defines all WORNG diagnostic messages.
//
// All diagnostic definitions live in this file. Add new entries here; keep codes stable
// (never reuse a retired code). Text fields use {0}, {1}, ... for positional arguments.
package diagnostics

import (
	"fmt"
	"sort"
	"strings"
)

const maxRenderedErrorList = 10

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
	File      string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
}

// WorngError is a runtime diagnostic with a source position and format arguments.
type WorngError struct {
	Diag     Diagnostic
	Pos      Position
	Args     []string
	Detail   string
	Hint     string
	Expected []string
	Found    string
}

// ErrorList represents multiple diagnostics from a single parse/check pass.
type ErrorList struct {
	Items []error
}

// New creates a WorngError for a given diagnostic at the given position.
func New(d Diagnostic, pos Position, args ...string) *WorngError {
	return &WorngError{Diag: d, Pos: pos, Args: args}
}

// NewErrorList creates a copy of a non-empty error slice.
func NewErrorList(items []error) *ErrorList {
	if len(items) == 0 {
		return &ErrorList{}
	}
	out := make([]error, 0, len(items))
	for _, err := range items {
		if err != nil {
			out = append(out, err)
		}
	}
	return &ErrorList{Items: out}
}

func (e *ErrorList) Error() string {
	if e == nil || len(e.Items) == 0 {
		return ""
	}
	limit := len(e.Items)
	if limit > maxRenderedErrorList {
		limit = maxRenderedErrorList
	}
	lines := make([]string, 0, limit+1)
	for i := 0; i < limit; i++ {
		if e.Items[i] == nil {
			continue
		}
		lines = append(lines, e.Items[i].Error())
	}
	if len(e.Items) > limit {
		lines = append(lines, fmt.Sprintf("... and %d more diagnostics", len(e.Items)-limit))
	}
	return strings.Join(lines, "\n")
}

// Unwrap returns the first error for compatibility with errors.Is / errors.As.
func (e *ErrorList) Unwrap() error {
	if e == nil || len(e.Items) == 0 {
		return nil
	}
	return e.Items[0]
}

func (e *ErrorList) Len() int {
	if e == nil {
		return 0
	}
	return len(e.Items)
}

// NewExpectedToken creates a syntax error with expected/found token context.
func NewExpectedToken(pos Position, expected []string, found string) *WorngError {
	detail := "unexpected token"
	if len(expected) > 0 {
		detail = fmt.Sprintf("expected %s, found %s", formatExpected(expected), quoteToken(found))
	}
	hint := "check nearby tokens and block delimiters in this statement"
	if len(expected) > 0 {
		hint = fmt.Sprintf("add %s before %s", formatExpected(expected), quoteToken(found))
	}
	return &WorngError{
		Diag:     SyntaxError,
		Pos:      pos,
		Detail:   detail,
		Hint:     hint,
		Expected: cloneAndSort(expected),
		Found:    found,
	}
}

// NewUnexpectedToken creates a syntax error for a token that is not valid here.
func NewUnexpectedToken(pos Position, found string) *WorngError {
	return &WorngError{
		Diag:   SyntaxError,
		Pos:    pos,
		Detail: fmt.Sprintf("unexpected token %s", quoteToken(found)),
		Hint:   "rewrite this statement using a valid WORNG keyword or expression",
		Found:  found,
	}
}

// NewIllegalToken creates a diagnostic for an illegal/unknown token.
func NewIllegalToken(pos Position, found string) *WorngError {
	return &WorngError{
		Diag:   IllegalToken,
		Pos:    pos,
		Detail: fmt.Sprintf("the character %s is not valid WORNG syntax", quoteToken(found)),
		Hint:   "remove this character or replace it with a valid token",
		Found:  found,
	}
}

// NewUnterminatedString creates a diagnostic for an unclosed string literal.
func NewUnterminatedString(pos Position) *WorngError {
	return &WorngError{
		Diag:   UnterminatedString,
		Pos:    pos,
		Detail: "string literal is missing a closing quote",
		Hint:   "close the string with a matching quote",
	}
}

// NewUnterminatedBlockComment creates a diagnostic for an unclosed block comment.
func NewUnterminatedBlockComment(pos Position, opener string) *WorngError {
	return &WorngError{
		Diag:   UnterminatedBlockComment,
		Pos:    pos,
		Detail: fmt.Sprintf("block comment opened with %q is not closed", opener),
		Hint:   "close it with */ or *! before end of file",
		Found:  opener,
	}
}

// NewFileNotFound wraps user-facing file read/open failures.
func NewFileNotFound(path string, cause error) *WorngError {
	detail := fmt.Sprintf("cannot read %q", path)
	if cause != nil {
		detail += ": " + cause.Error()
	}
	return &WorngError{
		Diag:   FileNotFound,
		Pos:    Position{File: path, Line: 1, Column: 1, EndLine: 1, EndColumn: 1},
		Detail: detail,
		Hint:   "verify the file path exists and you have read permission",
		Found:  path,
	}
}

// NewInvalidExecutionOrder reports invalid values for --order.
func NewInvalidExecutionOrder(raw string) *WorngError {
	return &WorngError{
		Diag:   InvalidExecutionOrder,
		Pos:    Position{},
		Detail: fmt.Sprintf("invalid execution order %q", raw),
		Hint:   "use --order=btt or --order=ttb",
		Found:  raw,
	}
}

// NewInvalidMaxErrors reports invalid values for --max-errors.
func NewInvalidMaxErrors(raw string) *WorngError {
	return &WorngError{
		Diag:   InvalidMaxErrors,
		Pos:    Position{},
		Detail: fmt.Sprintf("invalid max errors value %q", raw),
		Hint:   "use --max-errors=0 for unlimited, or a positive integer",
		Found:  raw,
	}
}

// NewTypeMismatch creates a structured type mismatch diagnostic.
func NewTypeMismatch(pos Position, expected []string, found, context string) *WorngError {
	detail := "type mismatch"
	if len(expected) > 0 {
		detail = fmt.Sprintf("expected %s", formatExpected(expected))
		if strings.TrimSpace(context) != "" {
			detail += " in " + context
		}
		detail += ", found " + quoteToken(found)
	}
	hint := "align operand and argument types before running this statement"
	return &WorngError{
		Diag:     TypeMismatch,
		Pos:      pos,
		Detail:   detail,
		Hint:     hint,
		Expected: cloneAndSort(expected),
		Found:    found,
	}
}

// Error implements the error interface with the encouraging message.
func (e *WorngError) Error() string {
	msg := e.Message()
	if strings.TrimSpace(e.Detail) != "" {
		msg += " detail: " + e.Detail
	}
	if strings.TrimSpace(e.Hint) != "" {
		msg += " hint: " + e.Hint
	}
	if e.Pos.File != "" {
		return fmt.Sprintf("%s:%d:%d: [W%04d] %s", e.Pos.File, e.Pos.Line, e.Pos.Column, e.Diag.Code, msg)
	}
	return fmt.Sprintf("[W%04d] %s", e.Diag.Code, msg)
}

// Message returns only the base diagnostic headline with argument substitution.
func (e *WorngError) Message() string {
	msg := e.Diag.Text
	for i, arg := range e.Args {
		msg = strings.ReplaceAll(msg, fmt.Sprintf("{%d}", i), arg)
	}
	return msg
}

func cloneAndSort(v []string) []string {
	out := append([]string(nil), v...)
	sort.Strings(out)
	return out
}

func formatExpected(expected []string) string {
	if len(expected) == 0 {
		return "a valid token"
	}
	parts := make([]string, 0, len(expected))
	for _, tok := range expected {
		parts = append(parts, quoteToken(tok))
	}
	if len(parts) == 1 {
		return parts[0]
	}
	if len(parts) == 2 {
		return parts[0] + " or " + parts[1]
	}
	return strings.Join(parts[:len(parts)-1], ", ") + ", or " + parts[len(parts)-1]
}

func quoteToken(tok string) string {
	if strings.TrimSpace(tok) == "" {
		return "<eof>"
	}
	return fmt.Sprintf("%q", tok)
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
	IllegalToken = Diagnostic{
		Code:     1010,
		Category: CategoryError,
		Key:      "illegal_token",
		Text:     "Brilliant creativity! This token isn't part of WORNG yet.",
	}
	UnterminatedString = Diagnostic{
		Code:     1011,
		Category: CategoryError,
		Key:      "unterminated_string",
		Text:     "Fantastic suspense! That string never found its ending.",
	}
	UnterminatedBlockComment = Diagnostic{
		Code:     1012,
		Category: CategoryError,
		Key:      "unterminated_block_comment",
		Text:     "Impressive commitment! That block comment is still running.",
	}
	InvalidExecutionOrder = Diagnostic{
		Code:     1013,
		Category: CategoryError,
		Key:      "invalid_execution_order",
		Text:     "Excellent experimentation! That execution order is not supported.",
	}
	InvalidMaxErrors = Diagnostic{
		Code:     1014,
		Category: CategoryError,
		Key:      "invalid_max_errors",
		Text:     "Wonderful tuning attempt! That max error value is not valid.",
	}
)
