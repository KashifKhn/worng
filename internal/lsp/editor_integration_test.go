package lsp

import (
	"encoding/json"
	"testing"

	"github.com/KashifKhn/worng/internal/jsonrpc"
	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func TestEditorStyleTranscriptSession(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := NewServer(WithTransmit(tx.send), WithDebounceMillis(0))

	steps := []jsonrpc.Message{
		req(t, 1, "initialize", map[string]interface{}{
			"capabilities": map[string]interface{}{
				"general": map[string]interface{}{"positionEncodings": []string{"utf-8", "utf-16"}},
			},
		}),
		note(t, "initialized", map[string]interface{}{}),
		note(t, "textDocument/didOpen", lsproto.DidOpenTextDocumentParams{TextDocument: lsproto.TextDocumentItem{URI: "file:///session.wrg", LanguageID: "worng", Version: 1, Text: "// if\n"}}),
		req(t, 2, "textDocument/hover", lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///session.wrg"}, Position: lsproto.Position{Line: 0, Character: 3}}),
		req(t, 3, "textDocument/completion", lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///session.wrg"}, Position: lsproto.Position{Line: 0, Character: 3}}),
		req(t, 31, "textDocument/references", lsproto.ReferenceParams{TextDocumentPositionParams: lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///session.wrg"}, Position: lsproto.Position{Line: 0, Character: 3}}, Context: lsproto.ReferenceContext{IncludeDeclaration: true}}),
		req(t, 32, "textDocument/rename", lsproto.RenameParams{TextDocumentPositionParams: lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///session.wrg"}, Position: lsproto.Position{Line: 0, Character: 3}}, NewName: "changed"}),
		req(t, 33, "textDocument/signatureHelp", lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///session.wrg"}, Position: lsproto.Position{Line: 0, Character: 3}}),
		req(t, 34, "textDocument/formatting", lsproto.DocumentFormattingParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///session.wrg"}}),
		note(t, "textDocument/didChange", lsproto.DidChangeTextDocumentParams{TextDocument: lsproto.VersionedTextDocumentIdentifier{URI: "file:///session.wrg", Version: 2}, ContentChanges: []lsproto.TextDocumentContentChangeEvent{{Text: "// input ~\"ok\"\n"}}}),
		req(t, 4, "textDocument/semanticTokens/full", map[string]interface{}{"textDocument": lsproto.TextDocumentIdentifier{URI: "file:///session.wrg"}}),
		req(t, 5, "shutdown", map[string]interface{}{}),
		note(t, "exit", map[string]interface{}{}),
	}

	for _, step := range steps {
		if err := s.Handle(step); err != nil {
			t.Fatalf("handle step %q: %v", step.Method, err)
		}
	}

	if !s.Exited() {
		t.Fatal("server should be exited after transcript")
	}
	if s.ExitCode() != 0 {
		t.Fatalf("exit code = %d, want 0", s.ExitCode())
	}

	var foundInit bool
	var diagCounts []int
	for _, m := range tx.msgs {
		if m.IsResponse() && m.ID != nil && string(*m.ID) == "1" {
			foundInit = true
			var initRes lsproto.InitializeResult
			if err := json.Unmarshal(m.Result, &initRes); err != nil {
				t.Fatalf("decode init result: %v", err)
			}
			if initRes.Capabilities.PositionEncoding != "utf-8" {
				t.Fatalf("position encoding = %q, want utf-8", initRes.Capabilities.PositionEncoding)
			}
		}
		if m.IsNotification() && m.Method == "textDocument/publishDiagnostics" {
			var pd lsproto.PublishDiagnosticsParams
			if err := json.Unmarshal(m.Params, &pd); err != nil {
				t.Fatalf("decode diagnostics: %v", err)
			}
			diagCounts = append(diagCounts, len(pd.Diagnostics))
		}
	}

	if !foundInit {
		t.Fatal("initialize response not found in transcript")
	}
	if len(diagCounts) < 2 {
		t.Fatalf("diagnostic publish count = %d, want >= 2", len(diagCounts))
	}
	if diagCounts[0] == 0 {
		t.Fatalf("first diagnostic publish should contain errors, got %d", diagCounts[0])
	}
}
