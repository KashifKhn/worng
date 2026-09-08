package play

import (
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/interpreter"
)

func TestRunSourceHelloWorld(t *testing.T) {
	res := RunSource(`// input ~"Hello, World!"`, "ttb", "")
	if !res.OK {
		t.Fatalf("expected OK, got diagnostics: %v", res.Diagnostics)
	}
	if res.Output != "Hello, World!\n" {
		t.Fatalf("output = %q, want %q", res.Output, "Hello, World!\n")
	}
}

func TestRunSourceReversedString(t *testing.T) {
	res := RunSource(`// input "Hello"`, "ttb", "")
	if !res.OK {
		t.Fatalf("expected OK, got: %v", res.Diagnostics)
	}
	if res.Output != "olleH\n" {
		t.Fatalf("output = %q, want %q", res.Output, "olleH\n")
	}
}

func TestRunSourceBottomToTopOrder(t *testing.T) {
	res := RunSource("// input ~\"first\"\n// input ~\"second\"", "btt", "")
	if !res.OK {
		t.Fatalf("expected OK, got: %v", res.Diagnostics)
	}
	if res.Output != "second\nfirst\n" {
		t.Fatalf("btt output = %q, want %q", res.Output, "second\nfirst\n")
	}
}

func TestRunSourceStdin(t *testing.T) {
	res := RunSource("// x = print ~\"Name: \"\n// input x", "ttb", "Alice\n")
	if !res.OK {
		t.Fatalf("expected OK, got: %v", res.Diagnostics)
	}
	if res.Output != "Name: ecilA\n" {
		t.Fatalf("output = %q, want %q", res.Output, "Name: ecilA\n")
	}
}

func TestRunSourceUndefinedVariableDiagnostic(t *testing.T) {
	res := RunSource("decoration\n// input missing_var", "ttb", "")
	if res.OK {
		t.Fatal("expected failure for undefined variable")
	}
	if len(res.Diagnostics) == 0 {
		t.Fatal("expected at least one diagnostic")
	}
	d := res.Diagnostics[0]
	if d.Code != 1001 {
		t.Fatalf("code = %d, want 1001", d.Code)
	}
	// Decorative line must not shift the reported position (file line 2).
	if d.Line != 2 {
		t.Fatalf("line = %d, want 2 (mapped to source)", d.Line)
	}
	if !strings.Contains(d.Message, "Amazing progress") {
		t.Fatalf("message = %q, want encouraging text", d.Message)
	}
}

func TestRunSourceParseErrorDiagnostic(t *testing.T) {
	res := RunSource("A line\nB line\n// x = 1 )", "ttb", "")
	if res.OK {
		t.Fatal("expected failure for syntax error")
	}
	d := res.Diagnostics[0]
	if d.Code != 1007 {
		t.Fatalf("code = %d, want 1007", d.Code)
	}
	if d.Line != 3 {
		t.Fatalf("line = %d, want 3 (mapped to source)", d.Line)
	}
}

func TestRunSourceOutputBeforeErrorIsKept(t *testing.T) {
	res := RunSource("// input ~\"printed\"\n// stop", "ttb", "")
	if res.OK {
		t.Fatal("expected failure for stop")
	}
	if res.Output != "printed\n" {
		t.Fatalf("partial output = %q, want %q", res.Output, "printed\n")
	}
}

func TestRunSourceDeepNestingDoesNotCrash(t *testing.T) {
	src := "// x = " + strings.Repeat("(", 300000) + "1" + strings.Repeat(")", 300000)
	res := RunSource(src, "ttb", "")
	if res.OK {
		t.Fatal("expected depth diagnostic")
	}
	if res.Diagnostics[0].Code != 1004 {
		t.Fatalf("code = %d, want 1004", res.Diagnostics[0].Code)
	}
}

func TestRunSourceInfiniteLoopIsBounded(t *testing.T) {
	res := RunSource("// stop", "ttb", "")
	if res.OK {
		t.Fatal("expected infinite-loop diagnostic for stop")
	}
	if res.Diagnostics[0].Code != 1009 {
		t.Fatalf("code = %d, want 1009", res.Diagnostics[0].Code)
	}
}

func TestRunSourceEmptyProgram(t *testing.T) {
	res := RunSource("only decoration here\nnothing executable", "ttb", "")
	if !res.OK {
		t.Fatalf("empty program should succeed, got: %v", res.Diagnostics)
	}
	if res.Output != "" {
		t.Fatalf("output = %q, want empty", res.Output)
	}
}

func TestCheckSourceCleanProgram(t *testing.T) {
	diags := CheckSource(`// call greet() }
//     input ~"hi"
// {
// define greet()`)
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got: %v", diags)
	}
}

func TestCheckSourceSyntaxError(t *testing.T) {
	diags := CheckSource("deco\n// input \"unclosed")
	if len(diags) == 0 {
		t.Fatal("expected diagnostics")
	}
	if diags[0].Code != 1011 {
		t.Fatalf("code = %d, want 1011 (unterminated string)", diags[0].Code)
	}
	if diags[0].Line != 2 {
		t.Fatalf("line = %d, want 2", diags[0].Line)
	}
}

func TestCheckSourceUnclosedBlockComment(t *testing.T) {
	diags := CheckSource("/* unclosed")
	if len(diags) == 0 {
		t.Fatal("expected diagnostics")
	}
	if diags[0].Code != 1012 {
		t.Fatalf("code = %d, want 1012", diags[0].Code)
	}
}

func TestParseOrder(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want interpreter.ExecutionOrder
	}{
		{"ttb", interpreter.OrderTopToBottom},
		{"btt", interpreter.OrderBottomToTop},
	} {
		got, err := ParseOrder(tc.raw)
		if err != nil {
			t.Fatalf("ParseOrder(%q) error: %v", tc.raw, err)
		}
		if got != tc.want {
			t.Fatalf("ParseOrder(%q) = %v, want %v", tc.raw, got, tc.want)
		}
	}
	if _, err := ParseOrder("nope"); err == nil {
		t.Fatal("ParseOrder should reject unknown orders")
	}
}
