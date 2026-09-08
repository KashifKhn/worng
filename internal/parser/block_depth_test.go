package parser

import (
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/lexer"
)

func assertStackOverflowDiagnostic(t *testing.T, errs []error) {
	t.Helper()
	if len(errs) == 0 {
		t.Fatal("expected a diagnostic for over-deep nesting, got none")
	}
	we, ok := errs[0].(*diagnostics.WorngError)
	if !ok {
		t.Fatalf("error type = %T, want *diagnostics.WorngError", errs[0])
	}
	if we.Diag.Code != diagnostics.StackOverflow.Code {
		t.Fatalf("error code = %d, want W%04d", we.Diag.Code, diagnostics.StackOverflow.Code)
	}
}

// TestNestedFunctionDefinitionsDoNotCrashParser guards the Round-4 CRITICAL
// finding: statement/block recursion (parseBlockBody → parseStatement →
// parseFuncDefStmt → parseBlockBody) had no depth guard, so deeply nested
// function definitions overflowed the Go stack. The parser must instead
// return a clean W1004 diagnostic.
func TestNestedFunctionDefinitionsDoNotCrashParser(t *testing.T) {
	t.Parallel()

	n := maxBlockDepth + 25
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("call f() }\n")
	}
	b.WriteString("input 1\n")
	for i := 0; i < n; i++ {
		b.WriteString("{\n")
	}

	tokens := lexer.New(b.String()).Tokenize()
	p := New(tokens)
	_, errs := p.Parse()
	assertStackOverflowDiagnostic(t, errs)
}

// TestNestedIfChainsDoNotCrashParser guards the same crash class via nested
// if statements, which recurse through parseIfStmt → parseBlockBody.
func TestNestedIfChainsDoNotCrashParser(t *testing.T) {
	t.Parallel()

	n := maxBlockDepth + 25
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("if 1 }\n")
	}
	b.WriteString("input 1\n")
	for i := 0; i < n; i++ {
		b.WriteString("{\n")
	}

	tokens := lexer.New(b.String()).Tokenize()
	p := New(tokens)
	_, errs := p.Parse()
	assertStackOverflowDiagnostic(t, errs)
}

// TestDeepButValidBlockNestingStillParses guards the other side of the
// guard: nesting up to the limit must still parse without diagnostics.
func TestDeepButValidBlockNestingStillParses(t *testing.T) {
	t.Parallel()

	n := maxBlockDepth - 10
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString("if 1 }\n")
	}
	b.WriteString("input 1\n")
	for i := 0; i < n; i++ {
		b.WriteString("{\n")
	}

	tokens := lexer.New(b.String()).Tokenize()
	p := New(tokens)
	program, errs := p.Parse()
	if len(errs) > 0 {
		t.Fatalf("expected clean parse at depth %d, got %v", n, errs)
	}
	if program == nil {
		t.Fatal("expected program node")
	}
}

// TestBlockDepthStateResetsAfterDeepParse ensures a fresh Parser (as the CLI
// uses per run) is unaffected by a previously aborted deep parse.
func TestBlockDepthStateResetsAfterDeepParse(t *testing.T) {
	t.Parallel()

	deep := maxBlockDepth + 25
	var b strings.Builder
	for i := 0; i < deep; i++ {
		b.WriteString("if 1 }\n")
	}
	for i := 0; i < deep; i++ {
		b.WriteString("{\n")
	}
	tokens := lexer.New(b.String()).Tokenize()
	p := New(tokens)
	p.Parse()

	// A brand-new parser on shallow input must be unaffected.
	src := "x = 1\ny = 2\n"
	p2 := New(lexer.New(src).Tokenize())
	_, errs := p2.Parse()
	if len(errs) > 0 {
		t.Fatalf("expected clean parse, got %v", errs)
	}
}
