package lsp

import (
	"testing"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func TestSignatureHelpBranches(t *testing.T) {
	t.Parallel()

	s := NewServer()
	if got := s.signatureHelp(lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///none"}}); got != nil {
		t.Fatalf("missing doc signature = %#v, want nil", got)
	}

	s.docs["file:///a"] = &document{uri: "file:///a", text: "", version: 1}
	if got := s.signatureHelp(lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a"}, Position: lsproto.Position{Line: 0, Character: 0}}); got != nil {
		t.Fatalf("empty line signature = %#v, want nil", got)
	}

	s.docs["file:///b"] = &document{uri: "file:///b", text: "// define add(1,2", version: 1}
	s.reindexDoc("file:///b", "// call add(a,b) }\n// discard a\n// {\n")
	got := s.signatureHelp(lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///b"}, Position: lsproto.Position{Line: 0, Character: 999}})
	if got == nil {
		t.Fatal("expected signature help")
	}
	if got.ActiveParameter < 0 {
		t.Fatalf("active parameter = %d", got.ActiveParameter)
	}
}
