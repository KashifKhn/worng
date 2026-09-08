package main

import (
	"os"
	"testing"

	"github.com/KashifKhn/worng/internal/vfs"
)

// TestFormatFilePreservesExecutableMarkers guards the Round-4 HIGH finding:
// worng fmt used lexer.Preprocess (an extraction pass) and wrote back only
// the excerpted contents, stripping the ///!! markers that make lines
// executable. The formatted program then silently ran as plain text. A
// formatter must preserve semantics: markers stay, only content is trimmed.
func TestFormatFilePreservesExecutableMarkers(t *testing.T) {
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
	want := "plain\n// x = 1\n!! input x\n"
	if string(got) != want {
		t.Fatalf("formatted = %q, want %q", string(got), want)
	}
}

// TestFormatFileOutputStillRuns confirms fmt is semantics-preserving end to
// end: the same executable lines are extracted before and after formatting.
// Internal spacing is preserved — collapsing it would corrupt string
// literals.
func TestFormatFileOutputStillRuns(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := "//   x = 5\n//   input x\n!!input ~\"ok\"\n"
	mustWriteProgram(t, fs, "fmt.wrg", source)

	if err := formatFile(fs, "fmt.wrg"); err != nil {
		t.Fatalf("formatFile error: %v", err)
	}

	got, err := fs.ReadFile("fmt.wrg")
	if err != nil {
		t.Fatalf("read formatted file: %v", err)
	}

	want := "// x = 5\n// input x\n!! input ~\"ok\"\n"
	if string(got) != want {
		t.Fatalf("formatted = %q, want %q", string(got), want)
	}
}

// TestFormatFileKeepsBlockCommentStructure verifies block-comment lines are
// normalized without losing the open/close markers that make their content
// executable.
func TestFormatFileKeepsBlockCommentStructure(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := "text\n/*\n   x = 1  \n   y = 2\n*/\ntail text\n"
	mustWriteProgram(t, fs, "blk.wrg", source)

	if err := formatFile(fs, "blk.wrg"); err != nil {
		t.Fatalf("formatFile error: %v", err)
	}

	got, err := fs.ReadFile("blk.wrg")
	if err != nil {
		t.Fatalf("read formatted file: %v", err)
	}
	want := "text\n/*\nx = 1\ny = 2\n*/\ntail text\n"
	if string(got) != want {
		t.Fatalf("formatted = %q, want %q", string(got), want)
	}
}

func TestFormatFileSingleLineBlockComment(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := "/*   x = 1   */\n!*  y = 2   *!\nplain\n"
	mustWriteProgram(t, fs, "one.wrg", source)

	if err := formatFile(fs, "one.wrg"); err != nil {
		t.Fatalf("formatFile error: %v", err)
	}

	got, err := fs.ReadFile("one.wrg")
	if err != nil {
		t.Fatalf("read formatted file: %v", err)
	}
	want := "/* x = 1 */\n!* y = 2 *!\nplain\n"
	if string(got) != want {
		t.Fatalf("formatted = %q, want %q", string(got), want)
	}
}

func TestFormatFileBareMarkerKeptWithoutTrailingSpace(t *testing.T) {
	t.Parallel()

	fs := vfs.NewMemFS()
	source := "//\n!!\n// x = 1\n"
	mustWriteProgram(t, fs, "bare.wrg", source)

	if err := formatFile(fs, "bare.wrg"); err != nil {
		t.Fatalf("formatFile error: %v", err)
	}

	got, err := fs.ReadFile("bare.wrg")
	if err != nil {
		t.Fatalf("read formatted file: %v", err)
	}
	want := "//\n!!\n// x = 1\n"
	if string(got) != want {
		t.Fatalf("formatted = %q, want %q", string(got), want)
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
