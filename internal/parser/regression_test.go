package parser

import (
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/lexer"
)

// TestParseUnterminatedStringInExpression covers an unterminated string that
// appears as a primary in an expression (e.g. after `input`), where parsePrimary
// previously reported a generic W1007 syntax error instead of W1011.
func TestParseUnterminatedStringInExpression(t *testing.T) {
	p := New(lexer.New(`input "abc`).Tokenize())
	_, errs := p.Parse()
	if len(errs) == 0 {
		t.Fatal("expected a parse error")
	}
	we, ok := errs[0].(*diagnostics.WorngError)
	if !ok {
		t.Fatalf("error type = %T, want *diagnostics.WorngError", errs[0])
	}
	if we.Diag.Code != diagnostics.UnterminatedString.Code {
		t.Fatalf("diag code = %d, want %d (%s)", we.Diag.Code, diagnostics.UnterminatedString.Code, errs[0].Error())
	}
}

// TestParseDeepNestingReturnsDiagnostic covers the CRITICAL crash where deeply
// nested parentheses or arrays made the recursive-descent parser overflow the
// Go stack. The parser must return a clean diagnostic instead of crashing.
func TestParseDeepNestingReturnsDiagnostic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "nested parentheses", input: "x = " + strings.Repeat("(", 186_931) + "1" + strings.Repeat(")", 186_931)},
		{name: "nested arrays", input: "x = " + strings.Repeat("[", 186_931) + strings.Repeat("]", 186_931)},
		{name: "nested unary minus", input: "x = " + strings.Repeat("-", 186_931) + "1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := New(lexer.New(tc.input).Tokenize())
			program, errs := p.Parse() // must not crash
			if len(errs) == 0 {
				t.Fatalf("expected a depth diagnostic, got none (program statements: %d)", len(program.Statements))
			}
			we, ok := errs[0].(*diagnostics.WorngError)
			if !ok {
				t.Fatalf("error type = %T, want *diagnostics.WorngError", errs[0])
			}
			if we.Diag.Code != diagnostics.SyntaxError.Code && we.Diag.Code != diagnostics.StackOverflow.Code {
				t.Fatalf("diag code = %d, want W1007 or W1004, got: %s", we.Diag.Code, errs[0].Error())
			}
		})
	}
}

// TestParseShallowNestingStillWorks guards against an over-aggressive depth
// limit rejecting legitimate programs.
func TestParseShallowNestingStillWorks(t *testing.T) {
	input := "x = " + strings.Repeat("(", 200) + "1" + strings.Repeat(")", 200)
	p := New(lexer.New(input).Tokenize())
	program, errs := p.Parse()
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(program.Statements) != 1 {
		t.Fatalf("statements = %d, want 1", len(program.Statements))
	}
}
