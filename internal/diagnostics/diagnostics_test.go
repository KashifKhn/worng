package diagnostics

import (
	"strings"
	"testing"
)

// ─── New ──────────────────────────────────────────────────────────────────────

func TestNewCreatesWorngError(t *testing.T) {
	t.Parallel()

	pos := Position{File: "foo.wrg", Line: 3, Column: 7}
	e := New(UndefinedVariable, pos, "myVar")

	if e.Diag.Code != UndefinedVariable.Code {
		t.Fatalf("Code = %d, want %d", e.Diag.Code, UndefinedVariable.Code)
	}
	if len(e.Args) != 1 || e.Args[0] != "myVar" {
		t.Fatalf("Args = %v, want [myVar]", e.Args)
	}
	if e.Pos != pos {
		t.Fatalf("Pos = %v, want %v", e.Pos, pos)
	}
}

func TestNewWithNoArgs(t *testing.T) {
	t.Parallel()

	e := New(DivisionByZero, Position{})
	if len(e.Args) != 0 {
		t.Fatalf("Args = %v, want empty", e.Args)
	}
}

func TestNewWithMultipleArgs(t *testing.T) {
	t.Parallel()

	e := New(TypeMismatch, Position{Line: 1, Column: 1}, "number", "string")
	if len(e.Args) != 2 {
		t.Fatalf("Args len = %d, want 2", len(e.Args))
	}
}

// ─── Error() formatting ───────────────────────────────────────────────────────

func TestErrorWithFilePosition(t *testing.T) {
	t.Parallel()

	pos := Position{File: "main.wrg", Line: 10, Column: 4}
	e := New(UndefinedVariable, pos, "x")
	msg := e.Error()

	if !strings.HasPrefix(msg, "main.wrg:10:4:") {
		t.Fatalf("Error() = %q, want prefix \"main.wrg:10:4:\"", msg)
	}
	if !strings.Contains(msg, "[W1001]") {
		t.Fatalf("Error() = %q, missing [W1001]", msg)
	}
	if !strings.Contains(msg, "x") {
		t.Fatalf("Error() = %q, missing arg substitution for 'x'", msg)
	}
}

func TestErrorWithoutFile(t *testing.T) {
	t.Parallel()

	e := New(DivisionByZero, Position{Line: 5, Column: 2})
	msg := e.Error()

	if !strings.HasPrefix(msg, "[W1003]") {
		t.Fatalf("Error() = %q, want prefix \"[W1003]\"", msg)
	}
	// Must NOT include file path prefix
	if strings.Contains(msg, ":5:2:") {
		t.Fatalf("Error() = %q, should not contain position when File is empty", msg)
	}
}

func TestErrorArgSubstitution(t *testing.T) {
	t.Parallel()

	e := New(UndefinedVariable, Position{}, "mySpecialVar")
	msg := e.Error()

	if !strings.Contains(msg, "mySpecialVar") {
		t.Fatalf("Error() = %q, want {0} replaced with mySpecialVar", msg)
	}
	if strings.Contains(msg, "{0}") {
		t.Fatalf("Error() = %q, placeholder {0} was not replaced", msg)
	}
}

func TestErrorNoArgSubstitutionPlaceholderLeftWhenNoArgs(t *testing.T) {
	t.Parallel()

	// UndefinedVariable has {0} in its text — with no args, {0} stays
	e := New(UndefinedVariable, Position{})
	msg := e.Error()
	// The raw placeholder remains because no arg was passed
	if !strings.Contains(msg, "{0}") {
		t.Fatalf("Error() = %q, expected {0} to remain when no args supplied", msg)
	}
}

func TestErrorImplementsErrorInterface(t *testing.T) {
	t.Parallel()

	var err error = New(SyntaxError, Position{})
	if err.Error() == "" {
		t.Fatalf("Error() returned empty string")
	}
}

// ─── All diagnostic codes are stable and unique ───────────────────────────────

func TestDiagnosticCodesAreUnique(t *testing.T) {
	t.Parallel()

	all := []Diagnostic{
		UndefinedVariable,
		TypeMismatch,
		DivisionByZero,
		StackOverflow,
		IndexOutOfBounds,
		ModuleNotFound,
		SyntaxError,
		FileNotFound,
		InfiniteLoop,
		IllegalToken,
		UnterminatedString,
		UnterminatedBlockComment,
		InvalidExecutionOrder,
		InvalidMaxErrors,
	}

	seen := map[int]string{}
	for _, d := range all {
		if prev, ok := seen[d.Code]; ok {
			t.Fatalf("duplicate code %d: %q and %q", d.Code, prev, d.Key)
		}
		seen[d.Code] = d.Key
	}
}

func TestDiagnosticKeysAreNonEmpty(t *testing.T) {
	t.Parallel()

	all := []Diagnostic{
		UndefinedVariable,
		TypeMismatch,
		DivisionByZero,
		StackOverflow,
		IndexOutOfBounds,
		ModuleNotFound,
		SyntaxError,
		FileNotFound,
		InfiniteLoop,
		IllegalToken,
		UnterminatedString,
		UnterminatedBlockComment,
		InvalidExecutionOrder,
		InvalidMaxErrors,
	}

	for _, d := range all {
		if d.Key == "" {
			t.Fatalf("diagnostic code %d has empty Key", d.Code)
		}
		if d.Text == "" {
			t.Fatalf("diagnostic %q has empty Text", d.Key)
		}
	}
}

func TestDiagnosticCodesMatchExpected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		d    Diagnostic
		code int
		key  string
	}{
		{UndefinedVariable, 1001, "undefined_variable"},
		{TypeMismatch, 1002, "type_mismatch"},
		{DivisionByZero, 1003, "division_by_zero"},
		{StackOverflow, 1004, "stack_overflow"},
		{IndexOutOfBounds, 1005, "index_out_of_bounds"},
		{ModuleNotFound, 1006, "module_not_found"},
		{SyntaxError, 1007, "syntax_error"},
		{FileNotFound, 1008, "file_not_found"},
		{InfiniteLoop, 1009, "infinite_loop"},
		{IllegalToken, 1010, "illegal_token"},
		{UnterminatedString, 1011, "unterminated_string"},
		{UnterminatedBlockComment, 1012, "unterminated_block_comment"},
		{InvalidExecutionOrder, 1013, "invalid_execution_order"},
		{InvalidMaxErrors, 1014, "invalid_max_errors"},
	}

	for _, tc := range tests {
		t.Run(tc.key, func(t *testing.T) {
			t.Parallel()
			if tc.d.Code != tc.code {
				t.Fatalf("Code = %d, want %d", tc.d.Code, tc.code)
			}
			if tc.d.Key != tc.key {
				t.Fatalf("Key = %q, want %q", tc.d.Key, tc.key)
			}
		})
	}
}

func TestAllDiagnosticsAreErrors(t *testing.T) {
	t.Parallel()

	all := []Diagnostic{
		UndefinedVariable, TypeMismatch, DivisionByZero, StackOverflow,
		IndexOutOfBounds, ModuleNotFound, SyntaxError, FileNotFound, InfiniteLoop, IllegalToken,
		UnterminatedString, UnterminatedBlockComment,
		InvalidExecutionOrder, InvalidMaxErrors,
	}

	for _, d := range all {
		if d.Category != CategoryError {
			t.Fatalf("diagnostic %q has category %v, want CategoryError", d.Key, d.Category)
		}
	}
}

// ─── Error() code formatting ──────────────────────────────────────────────────

func TestErrorCodeFormattedWithFourDigits(t *testing.T) {
	t.Parallel()

	// All defined codes are 1001-1009; formatted as W1001 .. W1009
	tests := []struct {
		d    Diagnostic
		want string
	}{
		{UndefinedVariable, "[W1001]"},
		{TypeMismatch, "[W1002]"},
		{DivisionByZero, "[W1003]"},
		{StackOverflow, "[W1004]"},
		{IndexOutOfBounds, "[W1005]"},
		{ModuleNotFound, "[W1006]"},
		{SyntaxError, "[W1007]"},
		{FileNotFound, "[W1008]"},
		{InfiniteLoop, "[W1009]"},
		{IllegalToken, "[W1010]"},
		{UnterminatedString, "[W1011]"},
		{UnterminatedBlockComment, "[W1012]"},
		{InvalidExecutionOrder, "[W1013]"},
		{InvalidMaxErrors, "[W1014]"},
	}

	for _, tc := range tests {
		t.Run(tc.d.Key, func(t *testing.T) {
			t.Parallel()
			e := New(tc.d, Position{})
			msg := e.Error()
			if !strings.Contains(msg, tc.want) {
				t.Fatalf("Error() = %q, want to contain %q", msg, tc.want)
			}
		})
	}
}

// ─── Position ─────────────────────────────────────────────────────────────────

func TestPositionZeroValue(t *testing.T) {
	t.Parallel()

	var p Position
	if p.File != "" || p.Line != 0 || p.Column != 0 || p.EndLine != 0 || p.EndColumn != 0 {
		t.Fatalf("zero Position = %+v, want all zero", p)
	}
}

func TestErrorIncludesDetailAndHint(t *testing.T) {
	t.Parallel()

	e := NewUnexpectedToken(Position{File: "main.wrg", Line: 2, Column: 5}, "}")
	msg := e.Error()

	if !strings.Contains(msg, "detail:") {
		t.Fatalf("Error() = %q, expected detail segment", msg)
	}
	if !strings.Contains(msg, "hint:") {
		t.Fatalf("Error() = %q, expected hint segment", msg)
	}
}

func TestNewExpectedTokenCarriesExpectedAndFound(t *testing.T) {
	t.Parallel()

	e := NewExpectedToken(Position{Line: 1, Column: 1}, []string{")", "identifier"}, "<eof>")
	if e.Diag.Code != SyntaxError.Code {
		t.Fatalf("Code = %d, want %d", e.Diag.Code, SyntaxError.Code)
	}
	if len(e.Expected) != 2 {
		t.Fatalf("Expected len = %d, want 2", len(e.Expected))
	}
	if e.Found != "<eof>" {
		t.Fatalf("Found = %q, want <eof>", e.Found)
	}
}

func TestErrorListRendersMultipleLines(t *testing.T) {
	t.Parallel()

	errList := NewErrorList([]error{
		New(SyntaxError, Position{File: "a.wrg", Line: 1, Column: 1}),
		New(IllegalToken, Position{File: "a.wrg", Line: 2, Column: 2}),
	})
	msg := errList.Error()
	if !strings.Contains(msg, "[W1007]") || !strings.Contains(msg, "[W1010]") {
		t.Fatalf("Error() = %q, expected both diagnostic codes", msg)
	}
	if !strings.Contains(msg, "\n") {
		t.Fatalf("Error() = %q, expected newline-separated diagnostics", msg)
	}
}

func TestNewUnterminatedConstructors(t *testing.T) {
	t.Parallel()

	str := NewUnterminatedString(Position{Line: 1, Column: 1})
	if str.Diag.Code != UnterminatedString.Code {
		t.Fatalf("code = %d, want %d", str.Diag.Code, UnterminatedString.Code)
	}
	blk := NewUnterminatedBlockComment(Position{Line: 1, Column: 1}, "/*")
	if blk.Diag.Code != UnterminatedBlockComment.Code {
		t.Fatalf("code = %d, want %d", blk.Diag.Code, UnterminatedBlockComment.Code)
	}
}

func TestNewFileNotFoundAndFlagDiagnostics(t *testing.T) {
	t.Parallel()

	fn := NewFileNotFound("missing.wrg", nil)
	if fn.Diag.Code != FileNotFound.Code {
		t.Fatalf("file not found code = %d, want %d", fn.Diag.Code, FileNotFound.Code)
	}
	ord := NewInvalidExecutionOrder("nope")
	if ord.Diag.Code != InvalidExecutionOrder.Code {
		t.Fatalf("invalid order code = %d, want %d", ord.Diag.Code, InvalidExecutionOrder.Code)
	}
	max := NewInvalidMaxErrors("abc")
	if max.Diag.Code != InvalidMaxErrors.Code {
		t.Fatalf("invalid max errors code = %d, want %d", max.Diag.Code, InvalidMaxErrors.Code)
	}
}
