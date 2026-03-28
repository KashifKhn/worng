package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/interpreter"
	"github.com/KashifKhn/worng/internal/vfs"
)

func TestRunFileExecutesBottomToTopViaPreprocessPipeline(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := strings.Join([]string{
		"plain text should be ignored",
		"// input ~\"first\"",
		"// input ~\"second\"",
		"",
	}, "\n")
	mustWriteProgram(t, fs, "program.wrg", source)

	var out bytes.Buffer
	if err := runFile(fs, "program.wrg", strings.NewReader(""), &out, interpreter.OrderBottomToTop, 20); err != nil {
		t.Fatalf("runFile error: %v", err)
	}

	if out.String() != "second\nfirst\n" {
		t.Fatalf("output = %q, want %q", out.String(), "second\nfirst\n")
	}
}

func TestRunFileMixedCommentStylesAllExecuteInBottomToTopOrder(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := strings.Join([]string{
		"// input ~\"line1\"",
		"!! input ~\"line2\"",
		"/*",
		"input ~\"line3\"",
		"*/",
		"!* input ~\"line4\" *!",
		"",
	}, "\n")
	mustWriteProgram(t, fs, "mixed.wrg", source)

	var out bytes.Buffer
	if err := runFile(fs, "mixed.wrg", strings.NewReader(""), &out, interpreter.OrderBottomToTop, 20); err != nil {
		t.Fatalf("runFile error: %v", err)
	}

	if out.String() != "line4\nline3\nline2\nline1\n" {
		t.Fatalf("output = %q, want %q", out.String(), "line4\nline3\nline2\nline1\n")
	}
}

func TestRunFileCRLFInputStillExecutesBottomToTop(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := "// input ~\"first\"\r\n!! input ~\"second\"\r\n"
	mustWriteProgram(t, fs, "crlf.wrg", source)

	var out bytes.Buffer
	if err := runFile(fs, "crlf.wrg", strings.NewReader(""), &out, interpreter.OrderBottomToTop, 20); err != nil {
		t.Fatalf("runFile error: %v", err)
	}

	if out.String() != "second\nfirst\n" {
		t.Fatalf("output = %q, want %q", out.String(), "second\nfirst\n")
	}
}

func TestRunFileIgnoresNonExecutableLines(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := strings.Join([]string{
		"input ~\"ignored\"",
		"x = 1",
		"plain text",
		"",
	}, "\n")
	mustWriteProgram(t, fs, "ignored.wrg", source)

	var out bytes.Buffer
	if err := runFile(fs, "ignored.wrg", strings.NewReader(""), &out, interpreter.OrderBottomToTop, 20); err != nil {
		t.Fatalf("runFile error: %v", err)
	}

	if out.String() != "" {
		t.Fatalf("output = %q, want empty", out.String())
	}
}

func TestRunFilePrintInputPipelineUsesRuntimeInversions(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := "// input print ~\"Name: \"\n"
	mustWriteProgram(t, fs, "io.wrg", source)

	var out bytes.Buffer
	if err := runFile(fs, "io.wrg", strings.NewReader("Alice\n"), &out, interpreter.OrderBottomToTop, 20); err != nil {
		t.Fatalf("runFile error: %v", err)
	}

	if out.String() != "Name: ecilA\n" {
		t.Fatalf("output = %q, want %q", out.String(), "Name: ecilA\n")
	}
}

func TestRunFileReturnsRuntimeDiagnostic(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	mustWriteProgram(t, fs, "stop.wrg", "// stop\n")

	err := runFile(fs, "stop.wrg", strings.NewReader(""), &bytes.Buffer{}, interpreter.OrderBottomToTop, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	we, ok := err.(*diagnostics.WorngError)
	if !ok {
		t.Fatalf("error type = %T, want *diagnostics.WorngError", err)
	}
	if we.Diag.Code != diagnostics.InfiniteLoop.Code {
		t.Fatalf("diag code = %d, want %d", we.Diag.Code, diagnostics.InfiniteLoop.Code)
	}
}

func TestRunFileReturnsParseDiagnostic(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	mustWriteProgram(t, fs, "bad.wrg", "// if\n")

	err := runFile(fs, "bad.wrg", strings.NewReader(""), &bytes.Buffer{}, interpreter.OrderBottomToTop, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errList, ok := err.(*diagnostics.ErrorList)
	if !ok {
		t.Fatalf("error type = %T, want *diagnostics.ErrorList", err)
	}
	if errList.Len() == 0 {
		t.Fatal("expected non-empty diagnostics list")
	}

	we, ok := errList.Unwrap().(*diagnostics.WorngError)
	if !ok {
		t.Fatalf("error type = %T, want *diagnostics.WorngError", errList.Unwrap())
	}
	if we.Diag.Code != diagnostics.SyntaxError.Code {
		t.Fatalf("diag code = %d, want %d", we.Diag.Code, diagnostics.SyntaxError.Code)
	}
	if we.Pos.File != "bad.wrg" {
		t.Fatalf("file = %q, want %q", we.Pos.File, "bad.wrg")
	}
}

func TestRunFileMissingFileReturnsPathError(t *testing.T) {
	t.Parallel()

	err := runFile(vfs.NewMemFS(), "missing.wrg", strings.NewReader(""), &bytes.Buffer{}, interpreter.OrderBottomToTop, 20)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	we, ok := err.(*diagnostics.WorngError)
	if !ok {
		t.Fatalf("error type = %T, want *diagnostics.WorngError", err)
	}
	if we.Diag.Code != diagnostics.FileNotFound.Code {
		t.Fatalf("diag code = %d, want %d", we.Diag.Code, diagnostics.FileNotFound.Code)
	}
}

func TestRunFileTopToBottomOrderOption(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := strings.Join([]string{
		"// input ~\"first\"",
		"// input ~\"second\"",
		"",
	}, "\n")
	mustWriteProgram(t, fs, "ttb.wrg", source)

	var out bytes.Buffer
	if err := runFile(fs, "ttb.wrg", strings.NewReader(""), &out, interpreter.OrderTopToBottom, 20); err != nil {
		t.Fatalf("runFile error: %v", err)
	}

	if out.String() != "first\nsecond\n" {
		t.Fatalf("output = %q, want %q", out.String(), "first\nsecond\n")
	}
}

func TestRunFileDefaultBottomToTopParsesNaturalIfElse(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := strings.Join([]string{
		"// input ~\"TAIL\"",
		"// if false }",
		"// input ~\"IF\"",
		"// { else }",
		"// input ~\"ELSE\"",
		"// {",
	}, "\n")
	mustWriteProgram(t, fs, "ifelse.wrg", source)

	var out bytes.Buffer
	if err := runFile(fs, "ifelse.wrg", strings.NewReader(""), &out, interpreter.OrderBottomToTop, 20); err != nil {
		t.Fatalf("runFile error: %v", err)
	}

	if out.String() != "ELSE\nTAIL\n" {
		t.Fatalf("output = %q, want %q", out.String(), "ELSE\nTAIL\n")
	}
}

func TestRunFileTopToBottomParsesNaturalIfElse(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := strings.Join([]string{
		"// input ~\"TAIL\"",
		"// if false }",
		"// input ~\"IF\"",
		"// { else }",
		"// input ~\"ELSE\"",
		"// {",
	}, "\n")
	mustWriteProgram(t, fs, "ifelse-ttb.wrg", source)

	var out bytes.Buffer
	if err := runFile(fs, "ifelse-ttb.wrg", strings.NewReader(""), &out, interpreter.OrderTopToBottom, 20); err != nil {
		t.Fatalf("runFile error: %v", err)
	}

	if out.String() != "TAIL\nELSE\n" {
		t.Fatalf("output = %q, want %q", out.String(), "TAIL\nELSE\n")
	}
}

func TestParseOrderFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		arg      string
		want     interpreter.ExecutionOrder
		consumed bool
		wantErr  bool
	}{
		{name: "no flag", arg: "program.wrg", consumed: false},
		{name: "btt", arg: "--order=btt", want: interpreter.OrderBottomToTop, consumed: true},
		{name: "ttb", arg: "--order=ttb", want: interpreter.OrderTopToBottom, consumed: true},
		{name: "invalid", arg: "--order=weird", consumed: true, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, consumed, err := parseOrderFlag(tc.arg)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseOrderFlag(%q) error: %v", tc.arg, err)
			}
			if consumed != tc.consumed {
				t.Fatalf("consumed = %v, want %v", consumed, tc.consumed)
			}
			if got != tc.want {
				t.Fatalf("order = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseExecutionFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		wantOrd  interpreter.ExecutionOrder
		wantJSON bool
		wantMax  int
		wantRest []string
		wantErr  bool
	}{
		{name: "no flags", args: []string{"prog.wrg"}, wantOrd: interpreter.OrderBottomToTop, wantMax: 20, wantRest: []string{"prog.wrg"}},
		{name: "json only", args: []string{"--json", "prog.wrg"}, wantOrd: interpreter.OrderBottomToTop, wantJSON: true, wantMax: 20, wantRest: []string{"prog.wrg"}},
		{name: "order and json", args: []string{"--order=ttb", "--json", "prog.wrg"}, wantOrd: interpreter.OrderTopToBottom, wantJSON: true, wantMax: 20, wantRest: []string{"prog.wrg"}},
		{name: "max errors", args: []string{"--max-errors=7", "prog.wrg"}, wantOrd: interpreter.OrderBottomToTop, wantMax: 7, wantRest: []string{"prog.wrg"}},
		{name: "invalid max errors", args: []string{"--max-errors=abc", "prog.wrg"}, wantErr: true},
		{name: "invalid order", args: []string{"--order=nope", "prog.wrg"}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			ord, jsonOut, maxErrs, rest, err := parseExecutionFlags(tc.args)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("parseExecutionFlags error: %v", err)
			}
			if ord != tc.wantOrd {
				t.Fatalf("order = %q, want %q", ord, tc.wantOrd)
			}
			if jsonOut != tc.wantJSON {
				t.Fatalf("json = %v, want %v", jsonOut, tc.wantJSON)
			}
			if maxErrs != tc.wantMax {
				t.Fatalf("max errors = %d, want %d", maxErrs, tc.wantMax)
			}
			if len(rest) != len(tc.wantRest) {
				t.Fatalf("rest len = %d, want %d", len(rest), len(tc.wantRest))
			}
			for i := range rest {
				if rest[i] != tc.wantRest[i] {
					t.Fatalf("rest[%d] = %q, want %q", i, rest[i], tc.wantRest[i])
				}
			}
		})
	}
}

func TestLimitErrors(t *testing.T) {
	t.Parallel()

	errA := diagnostics.New(diagnostics.SyntaxError, diagnostics.Position{})
	errB := diagnostics.New(diagnostics.SyntaxError, diagnostics.Position{})
	errC := diagnostics.New(diagnostics.SyntaxError, diagnostics.Position{})
	in := []error{errA, errB, errC}

	if got := limitErrors(in, 0); len(got) != 3 {
		t.Fatalf("len(limitErrors(0)) = %d, want 3", len(got))
	}
	if got := limitErrors(in, 2); len(got) != 2 {
		t.Fatalf("len(limitErrors(2)) = %d, want 2", len(got))
	}
}

func TestJoinExecutableLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{name: "empty", lines: nil, want: ""},
		{name: "single", lines: []string{"x = 1"}, want: "x = 1\n"},
		{name: "multiple", lines: []string{"a", "b"}, want: "a\nb\n"},
		{name: "includes blank", lines: []string{"", "x"}, want: "\nx\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := joinExecutableLines(tc.lines); got != tc.want {
				t.Fatalf("joinExecutableLines() = %q, want %q", got, tc.want)
			}
		})
	}
}

func mustWriteProgram(t *testing.T, fs vfs.FS, path, source string) {
	t.Helper()
	if err := fs.WriteFile(path, []byte(source)); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
