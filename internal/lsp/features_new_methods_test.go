package lsp

import (
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func TestReferences(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///a.wrg", "// x = 1\n// input x\n")
	reqP := lsproto.ReferenceParams{
		TextDocumentPositionParams: lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a.wrg"}, Position: lsproto.Position{Line: 1, Character: 9}},
		Context:                    lsproto.ReferenceContext{IncludeDeclaration: true},
	}
	if err := s.Handle(req(t, 200, "textDocument/references", reqP)); err != nil {
		t.Fatalf("references handle: %v", err)
	}
	resp := tx.lastResponse(t)
	var locs []lsproto.Location
	decodeResult(t, resp.Result, &locs)
	if len(locs) < 2 {
		t.Fatalf("references count = %d, want >= 2", len(locs))
	}
}

func TestRename(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///a.wrg", "// x = 1\n// input x\n")
	reqP := lsproto.RenameParams{
		TextDocumentPositionParams: lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a.wrg"}, Position: lsproto.Position{Line: 1, Character: 9}},
		NewName:                    "renamed",
	}
	if err := s.Handle(req(t, 201, "textDocument/rename", reqP)); err != nil {
		t.Fatalf("rename handle: %v", err)
	}
	resp := tx.lastResponse(t)
	var edit lsproto.WorkspaceEdit
	decodeResult(t, resp.Result, &edit)
	if len(edit.Changes) == 0 {
		t.Fatal("expected workspace edit changes")
	}
}

func TestSignatureHelp(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///a.wrg", "// call add(a,b) }\n// discard a\n// {\n// define add(1,\n")
	reqP := lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a.wrg"}, Position: lsproto.Position{Line: 3, Character: 15}}
	if err := s.Handle(req(t, 202, "textDocument/signatureHelp", reqP)); err != nil {
		t.Fatalf("signatureHelp handle: %v", err)
	}
	resp := tx.lastResponse(t)
	var help lsproto.SignatureHelp
	decodeResult(t, resp.Result, &help)
	if len(help.Signatures) == 0 {
		t.Fatal("expected at least one signature")
	}
}

func TestDocumentFormatting(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///a.wrg", "  // x = 1   \n\n")
	if err := s.Handle(req(t, 203, "textDocument/formatting", lsproto.DocumentFormattingParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a.wrg"}})); err != nil {
		t.Fatalf("formatting handle: %v", err)
	}
	resp := tx.lastResponse(t)
	var edits []lsproto.TextEdit
	decodeResult(t, resp.Result, &edits)
	if len(edits) == 0 {
		t.Fatal("expected formatting edits")
	}
	if !strings.Contains(edits[0].NewText, "// x = 1") {
		t.Fatalf("formatted text = %q", edits[0].NewText)
	}
}

func TestIncrementalDidChange(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///a.wrg", "// x = 1\n")
	rng := lsproto.Range{Start: lsproto.Position{Line: 0, Character: 3}, End: lsproto.Position{Line: 0, Character: 4}}
	chg := lsproto.DidChangeTextDocumentParams{
		TextDocument: lsproto.VersionedTextDocumentIdentifier{URI: "file:///a.wrg", Version: 2},
		ContentChanges: []lsproto.TextDocumentContentChangeEvent{{
			Range: &rng,
			Text:  "z",
		}},
	}
	if err := s.Handle(note(t, "textDocument/didChange", chg)); err != nil {
		t.Fatalf("incremental didChange: %v", err)
	}
	if s.docs["file:///a.wrg"].text != "// z = 1\n" {
		t.Fatalf("doc text = %q, want updated incremental text", s.docs["file:///a.wrg"].text)
	}
}

func TestCancelRequestPath(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(note(t, "$/cancelRequest", map[string]interface{}{"id": 10})); err != nil {
		t.Fatalf("cancel request handle: %v", err)
	}
	if !s.canceled["10"] {
		t.Fatal("expected canceled id recorded")
	}
}

func TestCanceledRequestReturnsCanceledError(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///a.wrg", "// input ~\"x\"\n")
	if err := s.Handle(note(t, "$/cancelRequest", map[string]interface{}{"id": 333})); err != nil {
		t.Fatalf("cancel note error: %v", err)
	}
	if err := s.Handle(req(t, 333, "textDocument/hover", lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a.wrg"}, Position: lsproto.Position{Line: 0, Character: 3}})); err != nil {
		t.Fatalf("hover request after cancel: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error == nil {
		t.Fatal("expected canceled error response")
	}
}

func TestReferencesExcludeDeclaration(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///a.wrg", "// x = 1\n// input x\n")
	reqP := lsproto.ReferenceParams{
		TextDocumentPositionParams: lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a.wrg"}, Position: lsproto.Position{Line: 1, Character: 9}},
		Context:                    lsproto.ReferenceContext{IncludeDeclaration: false},
	}
	if err := s.Handle(req(t, 204, "textDocument/references", reqP)); err != nil {
		t.Fatalf("references handle: %v", err)
	}
	resp := tx.lastResponse(t)
	var locs []lsproto.Location
	decodeResult(t, resp.Result, &locs)
	if len(locs) != 1 {
		t.Fatalf("references count = %d, want 1", len(locs))
	}
}

func TestReferencesNoWord(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///a.wrg", "// input ~\"x\"\n")
	reqP := lsproto.ReferenceParams{
		TextDocumentPositionParams: lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a.wrg"}, Position: lsproto.Position{Line: 0, Character: 0}},
		Context:                    lsproto.ReferenceContext{IncludeDeclaration: true},
	}
	if err := s.Handle(req(t, 205, "textDocument/references", reqP)); err != nil {
		t.Fatalf("references handle: %v", err)
	}
	resp := tx.lastResponse(t)
	if strings.TrimSpace(string(resp.Result)) != "null" && strings.TrimSpace(string(resp.Result)) != "[]" {
		t.Fatalf("references result = %s", string(resp.Result))
	}
}

func TestRenameEmptyNameReturnsNull(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///a.wrg", "// x = 1\n")
	reqP := lsproto.RenameParams{
		TextDocumentPositionParams: lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a.wrg"}, Position: lsproto.Position{Line: 0, Character: 3}},
		NewName:                    "   ",
	}
	if err := s.Handle(req(t, 206, "textDocument/rename", reqP)); err != nil {
		t.Fatalf("rename handle: %v", err)
	}
	resp := tx.lastResponse(t)
	if strings.TrimSpace(string(resp.Result)) != "null" {
		t.Fatalf("rename result = %s, want null", string(resp.Result))
	}
}

func TestSignatureHelpNoMatchReturnsNull(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///a.wrg", "// input ~\"x\"\n")
	reqP := lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a.wrg"}, Position: lsproto.Position{Line: 0, Character: 2}}
	if err := s.Handle(req(t, 207, "textDocument/signatureHelp", reqP)); err != nil {
		t.Fatalf("signatureHelp handle: %v", err)
	}
	resp := tx.lastResponse(t)
	if strings.TrimSpace(string(resp.Result)) != "null" {
		t.Fatalf("signature result = %s, want null", string(resp.Result))
	}
}

func TestFormattingMissingDocumentReturnsNull(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(req(t, 208, "textDocument/formatting", lsproto.DocumentFormattingParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///missing.wrg"}})); err != nil {
		t.Fatalf("formatting handle: %v", err)
	}
	resp := tx.lastResponse(t)
	if strings.TrimSpace(string(resp.Result)) != "null" {
		t.Fatalf("format result = %s, want null", string(resp.Result))
	}
}

func TestNewMethodsInvalidParams(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	methods := []string{"textDocument/references", "textDocument/rename", "textDocument/signatureHelp", "textDocument/formatting"}
	for i, m := range methods {
		if err := s.Handle(req(t, 300+i, m, map[string]interface{}{"x": 1})); err != nil {
			t.Fatalf("method %s handle: %v", m, err)
		}
		resp := tx.lastResponse(t)
		if resp.Error == nil {
			t.Fatalf("method %s expected error", m)
		}
	}
}
