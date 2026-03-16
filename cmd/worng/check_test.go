package main

import (
	"bytes"
	"os"
	"testing"

	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/interpreter"
	"github.com/KashifKhn/worng/internal/vfs"
)

func TestCheckFileSuccess(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	mustWriteProgram(t, fs, "ok.wrg", "plain\n// input ~\"ok\"\n")

	if err := checkFile(fs, "ok.wrg", interpreter.OrderBottomToTop); err != nil {
		t.Fatalf("checkFile error: %v", err)
	}
}

func TestCheckFileSyntaxError(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	mustWriteProgram(t, fs, "bad.wrg", "// if\n")

	err := checkFile(fs, "bad.wrg", interpreter.OrderTopToBottom)
	if err == nil {
		t.Fatal("expected syntax error, got nil")
	}

	we, ok := err.(*diagnostics.WorngError)
	if !ok {
		t.Fatalf("error type = %T, want *diagnostics.WorngError", err)
	}
	if we.Diag.Code != diagnostics.SyntaxError.Code {
		t.Fatalf("diag code = %d, want %d", we.Diag.Code, diagnostics.SyntaxError.Code)
	}
}

func TestCheckFileMissingFile(t *testing.T) {
	t.Parallel()

	err := checkFile(vfs.NewMemFS(), "missing.wrg", interpreter.OrderBottomToTop)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*os.PathError); !ok {
		t.Fatalf("error type = %T, want *os.PathError", err)
	}
}

func TestCheckCommandUsage(t *testing.T) {
	t.Parallel()

	if code := checkCommand(nil); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

func TestCheckCommandAcceptsOrderFlag(t *testing.T) {
	t.Parallel()

	if code := checkCommand([]string{"--order=ttb", "nonexistent.wrg"}); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}

func TestCheckCommandRejectsInvalidOrderFlag(t *testing.T) {
	t.Parallel()

	if code := checkCommand([]string{"--order=nope", "any.wrg"}); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

func TestRunCommandUsage(t *testing.T) {
	t.Parallel()

	if code := runCommand(nil); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

func TestFmtCommandUsage(t *testing.T) {
	t.Parallel()

	if code := fmtCommand(nil); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

func TestRunREPLEmptyInput(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := runREPL(bytes.NewBuffer(nil), &out, &errOut, interpreter.OrderBottomToTop)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if errOut.String() != "" {
		t.Fatalf("stderr = %q, want empty", errOut.String())
	}
	if out.String() != "WORNG v0.1.0 — Type // or !! followed by WORNG code.\n>>> " {
		t.Fatalf("stdout = %q", out.String())
	}
}
