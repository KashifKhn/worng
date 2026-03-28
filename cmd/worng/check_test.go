package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/interpreter"
	"github.com/KashifKhn/worng/internal/vfs"
)

func TestCheckFileSuccess(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	mustWriteProgram(t, fs, "ok.wrg", "plain\n// input ~\"ok\"\n")

	if err := checkFile(fs, "ok.wrg", interpreter.OrderBottomToTop, 20); err != nil {
		t.Fatalf("checkFile error: %v", err)
	}
}

func TestCheckFileSyntaxError(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	mustWriteProgram(t, fs, "bad.wrg", "// if\n")

	err := checkFile(fs, "bad.wrg", interpreter.OrderTopToBottom, 20)
	if err == nil {
		t.Fatal("expected syntax error, got nil")
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
		t.Fatalf("error type = %T, want *diagnostics.WorngError", err)
	}
	if we.Diag.Code != diagnostics.SyntaxError.Code {
		t.Fatalf("diag code = %d, want %d", we.Diag.Code, diagnostics.SyntaxError.Code)
	}
	if we.Pos.File != "bad.wrg" {
		t.Fatalf("file = %q, want %q", we.Pos.File, "bad.wrg")
	}
}

func TestCheckFileMissingFile(t *testing.T) {
	t.Parallel()

	err := checkFile(vfs.NewMemFS(), "missing.wrg", interpreter.OrderBottomToTop, 20)
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

func TestCheckCommandAcceptsJSONFlag(t *testing.T) {
	t.Parallel()

	if code := checkCommand([]string{"--json", "nonexistent.wrg"}); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}

func TestCheckCommandAcceptsMaxErrorsFlag(t *testing.T) {
	t.Parallel()

	if code := checkCommand([]string{"--max-errors=1", "nonexistent.wrg"}); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
}

func TestCheckCommandRejectsInvalidMaxErrorsFlag(t *testing.T) {
	t.Parallel()

	if code := checkCommand([]string{"--max-errors=abc", "anything.wrg"}); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
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

func TestCheckCommandJSONOutput(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}
	os.Stdout = wOut
	os.Stderr = wErr
	_ = rOut

	if err := os.WriteFile("tmp_bad_check.wrg", []byte("// if\n"), 0644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	defer func() {
		if err := os.Remove("tmp_bad_check.wrg"); err != nil {
			t.Fatalf("remove temp file: %v", err)
		}
	}()

	code := checkCommand([]string{"--json", "tmp_bad_check.wrg"})
	_ = wOut.Close()
	_ = wErr.Close()

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}

	var stderr bytes.Buffer
	if _, err := stderr.ReadFrom(rErr); err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	if !strings.Contains(stderr.String(), `"code": 1007`) {
		t.Fatalf("stderr = %q, want JSON diagnostics", stderr.String())
	}
}

func TestCheckCommandJSONOutputForInvalidMaxErrors(t *testing.T) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	rErr, wErr, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stderr: %v", err)
	}
	os.Stdout = wOut
	os.Stderr = wErr
	_ = rOut

	code := checkCommand([]string{"--json", "--max-errors=abc", "anything.wrg"})
	_ = wOut.Close()
	_ = wErr.Close()
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}

	var stderr bytes.Buffer
	if _, err := stderr.ReadFrom(rErr); err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	if !strings.Contains(stderr.String(), `"code": 1014`) {
		t.Fatalf("stderr = %q, want invalid max errors JSON diagnostic", stderr.String())
	}
}
