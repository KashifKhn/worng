package lsp

import (
	"testing"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func TestLookupDeclarationAndSameLocation(t *testing.T) {
	t.Parallel()

	s := NewServer()
	if got := s.lookupDeclaration("none"); got != nil {
		t.Fatalf("lookup none = %#v, want nil", got)
	}

	s.reindexDoc("file:///a", "// call f(a) }\n// discard a\n// {\n")
	loc := s.lookupDeclaration("f")
	if loc == nil {
		t.Fatal("expected declaration location")
	}
	if !sameLocation(*loc, *loc) {
		t.Fatal("sameLocation self should be true")
	}
}

func TestRenameNoWordAndCanceled(t *testing.T) {
	t.Parallel()

	s := NewServer()
	s.docs["file:///a"] = &document{uri: "file:///a", text: "// input ~\"x\"\n", version: 1}
	if edit := s.rename(lsproto.RenameParams{TextDocumentPositionParams: lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a"}, Position: lsproto.Position{Line: 0, Character: 0}}, NewName: "y"}); edit != nil {
		t.Fatalf("rename no-word edit = %#v, want nil", edit)
	}

	s.docs["file:///b"] = &document{uri: "file:///b", text: "// x = 1\n", version: 1}
	s.canceled["any"] = true
	if edit := s.rename(lsproto.RenameParams{TextDocumentPositionParams: lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///b"}, Position: lsproto.Position{Line: 0, Character: 3}}, NewName: "z"}); edit != nil {
		t.Fatalf("rename canceled edit = %#v, want nil", edit)
	}
}
