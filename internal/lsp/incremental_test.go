package lsp

import (
	"testing"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func TestApplyIncrementalChangesInvalidRangeIgnored(t *testing.T) {
	t.Parallel()

	text := "// x = 1\n"
	r := lsproto.Range{Start: lsproto.Position{Line: 100, Character: 0}, End: lsproto.Position{Line: 100, Character: 1}}
	out := applyIncrementalChanges(text, []lsproto.TextDocumentContentChangeEvent{{Range: &r, Text: "z"}})
	if out != text {
		t.Fatalf("text changed unexpectedly: %q", out)
	}
}

func TestPositionToOffsetBounds(t *testing.T) {
	t.Parallel()

	if got := positionToOffset("x", lsproto.Position{Line: -1, Character: 0}); got != -1 {
		t.Fatalf("line -1 offset = %d, want -1", got)
	}
	if got := positionToOffset("x", lsproto.Position{Line: 0, Character: -1}); got != -1 {
		t.Fatalf("char -1 offset = %d, want -1", got)
	}
	if got := positionToOffset("x", lsproto.Position{Line: 1, Character: 0}); got != -1 {
		t.Fatalf("line out of bounds offset = %d, want -1", got)
	}
	if got := positionToOffset("x", lsproto.Position{Line: 0, Character: 2}); got != -1 {
		t.Fatalf("char out of bounds offset = %d, want -1", got)
	}
}

func TestApplyIncrementalChangesFullReplaceFallback(t *testing.T) {
	t.Parallel()

	out := applyIncrementalChanges("old", []lsproto.TextDocumentContentChangeEvent{{Text: "new"}})
	if out != "new" {
		t.Fatalf("text = %q, want new", out)
	}
}
