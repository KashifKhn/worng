package lsp

import "testing"

func TestFormatDocument(t *testing.T) {
	t.Parallel()

	in := "  // x = 1   \n\n"
	out := formatDocument(in)
	if out != "// x = 1\n" {
		t.Fatalf("formatted = %q", out)
	}
}

func TestFullDocumentRange(t *testing.T) {
	t.Parallel()

	r := fullDocumentRange("a\n")
	if r.Start.Line != 0 || r.End.Line != 0 || r.End.Character != 1 {
		t.Fatalf("range = %#v", r)
	}
}

func TestFormatDocumentEmptyAndRangeEmpty(t *testing.T) {
	t.Parallel()

	if got := formatDocument(""); got != "" {
		t.Fatalf("format empty = %q", got)
	}
	r := fullDocumentRange("")
	if r.Start.Line != 0 || r.End.Line != 0 || r.End.Character != 0 {
		t.Fatalf("range empty = %#v", r)
	}
}
