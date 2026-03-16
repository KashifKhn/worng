package main

import (
	"os"
	"testing"

	"github.com/KashifKhn/worng/internal/vfs"
)

func TestFormatFileNormalizesExecutableLines(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := "plain\n//   x = 1   \n!!   input x\n"
	mustWriteProgram(t, fs, "fmt.wrg", source)

	if err := formatFile(fs, "fmt.wrg"); err != nil {
		t.Fatalf("formatFile error: %v", err)
	}

	got, err := fs.ReadFile("fmt.wrg")
	if err != nil {
		t.Fatalf("read formatted file: %v", err)
	}
	if string(got) != "x = 1\ninput x\n" {
		t.Fatalf("formatted = %q, want %q", string(got), "x = 1\ninput x\n")
	}
}

func TestFormatFileNoExecutableLinesWritesEmpty(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	mustWriteProgram(t, fs, "empty.wrg", "plain\ntext\n")

	if err := formatFile(fs, "empty.wrg"); err != nil {
		t.Fatalf("formatFile error: %v", err)
	}

	got, err := fs.ReadFile("empty.wrg")
	if err != nil {
		t.Fatalf("read formatted file: %v", err)
	}
	if string(got) != "" {
		t.Fatalf("formatted = %q, want empty", string(got))
	}
}

func TestFormatFileMissingFile(t *testing.T) {
	t.Parallel()

	err := formatFile(vfs.NewMemFS(), "missing.wrg")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*os.PathError); !ok {
		t.Fatalf("error type = %T, want *os.PathError", err)
	}
}
