package lsp

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/jsonrpc"
	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func TestDidChangeUnknownDocIgnored(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	change := lsproto.DidChangeTextDocumentParams{
		TextDocument:   lsproto.VersionedTextDocumentIdentifier{URI: "file:///missing.wrg", Version: 1},
		ContentChanges: []lsproto.TextDocumentContentChangeEvent{{Text: "// if\n"}},
	}
	if err := s.Handle(note(t, "textDocument/didChange", change)); err != nil {
		t.Fatalf("didChange error: %v", err)
	}
}

func TestDidChangeInvalidParamsIgnored(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	msg := jsonrpc.Message{JSONRPC: "2.0", Method: "textDocument/didChange", Params: json.RawMessage(`{"textDocument":true}`)}
	if err := s.Handle(msg); err != nil {
		t.Fatalf("didChange invalid params error: %v", err)
	}
}

func TestDidOpenInvalidParamsIgnored(t *testing.T) {
	t.Parallel()

	s := initialized(t, &capture{})
	msg := jsonrpc.Message{JSONRPC: "2.0", Method: "textDocument/didOpen", Params: json.RawMessage(`{"textDocument":1}`)}
	if err := s.Handle(msg); err != nil {
		t.Fatalf("didOpen invalid params error: %v", err)
	}
}

func TestDidCloseInvalidParamsIgnored(t *testing.T) {
	t.Parallel()

	s := initialized(t, &capture{})
	msg := jsonrpc.Message{JSONRPC: "2.0", Method: "textDocument/didClose", Params: json.RawMessage(`{"textDocument":1}`)}
	if err := s.Handle(msg); err != nil {
		t.Fatalf("didClose invalid params error: %v", err)
	}
}

func TestInvalidHoverParamsReturnsInvalidParams(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(jsonrpc.Message{JSONRPC: "2.0", ID: rawID(1), Method: "textDocument/hover", Params: json.RawMessage(`{"x":1}`)}); err != nil {
		t.Fatalf("hover invalid params error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error == nil || resp.Error.Code != jsonrpc.ErrorInvalidParams {
		t.Fatalf("error = %#v, want %d", resp.Error, jsonrpc.ErrorInvalidParams)
	}
}

func TestInvalidCompletionParamsReturnsInvalidParams(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(jsonrpc.Message{JSONRPC: "2.0", ID: rawID(2), Method: "textDocument/completion", Params: json.RawMessage(`{"x":1}`)}); err != nil {
		t.Fatalf("completion invalid params error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error == nil || resp.Error.Code != jsonrpc.ErrorInvalidParams {
		t.Fatalf("error = %#v, want %d", resp.Error, jsonrpc.ErrorInvalidParams)
	}
}

func TestInvalidDefinitionParamsReturnsInvalidParams(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(jsonrpc.Message{JSONRPC: "2.0", ID: rawID(3), Method: "textDocument/definition", Params: json.RawMessage(`{"x":1}`)}); err != nil {
		t.Fatalf("definition invalid params error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error == nil || resp.Error.Code != jsonrpc.ErrorInvalidParams {
		t.Fatalf("error = %#v, want %d", resp.Error, jsonrpc.ErrorInvalidParams)
	}
}

func TestInvalidDocumentSymbolParamsReturnsInvalidParams(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(jsonrpc.Message{JSONRPC: "2.0", ID: rawID(4), Method: "textDocument/documentSymbol", Params: json.RawMessage(`{"x":1}`)}); err != nil {
		t.Fatalf("documentSymbol invalid params error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error == nil || resp.Error.Code != jsonrpc.ErrorInvalidParams {
		t.Fatalf("error = %#v, want %d", resp.Error, jsonrpc.ErrorInvalidParams)
	}
}

func TestInvalidSemanticTokensParamsReturnsInvalidParams(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(jsonrpc.Message{JSONRPC: "2.0", ID: rawID(5), Method: "textDocument/semanticTokens/full", Params: json.RawMessage(`{"x":1}`)}); err != nil {
		t.Fatalf("semanticTokens invalid params error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error == nil || resp.Error.Code != jsonrpc.ErrorInvalidParams {
		t.Fatalf("error = %#v, want %d", resp.Error, jsonrpc.ErrorInvalidParams)
	}
}

func TestHandleResponseMessageIgnored(t *testing.T) {
	t.Parallel()

	s := NewServer()
	id := json.RawMessage("1")
	msg := jsonrpc.Message{JSONRPC: "2.0", ID: &id, Result: json.RawMessage(`null`)}
	if err := s.Handle(msg); err != nil {
		t.Fatalf("response handle error: %v", err)
	}
}

func TestHandleUnknownMessageShapeIgnored(t *testing.T) {
	t.Parallel()

	s := NewServer()
	msg := jsonrpc.Message{JSONRPC: "2.0"}
	if err := s.Handle(msg); err != nil {
		t.Fatalf("handle unknown shape: %v", err)
	}
}

func TestScheduleDiagnosticsWithDebounce(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := NewServer(WithTransmit(tx.send), WithDebounceMillis(1))
	if err := s.Handle(req(t, 1, "initialize", lsproto.InitializeParams{Capabilities: map[string]interface{}{}})); err != nil {
		t.Fatalf("initialize error: %v", err)
	}
	open := lsproto.DidOpenTextDocumentParams{TextDocument: lsproto.TextDocumentItem{URI: "file:///d.wrg", LanguageID: "worng", Version: 1, Text: "// if\n"}}
	if err := s.Handle(note(t, "textDocument/didOpen", open)); err != nil {
		t.Fatalf("didOpen error: %v", err)
	}
}

func TestPublishDiagnosticsDirect(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := NewServer(WithTransmit(tx.send))
	v := 1
	if err := s.publishDiagnostics("file:///x", &v, []lsproto.Diagnostic{}); err != nil {
		t.Fatalf("publishDiagnostics error: %v", err)
	}
	msg := tx.lastNotification(t, "textDocument/publishDiagnostics")
	if !strings.Contains(string(msg.Params), "file:///x") {
		t.Fatalf("publish params = %s", string(msg.Params))
	}
}

func TestWordAtBounds(t *testing.T) {
	t.Parallel()

	text := "alpha beta"
	word, _ := wordAt(text, lsproto.Position{Line: 0, Character: 100})
	if word != "beta" {
		t.Fatalf("word = %q, want beta", word)
	}
	word, _ = wordAt(text, lsproto.Position{Line: 0, Character: -1})
	if word != "alpha" {
		t.Fatalf("word = %q, want alpha", word)
	}
}

func TestIdentRangeBounds(t *testing.T) {
	t.Parallel()

	r := identRange(0, 0, 0)
	if r.Start.Line != 0 || r.Start.Character != 0 || r.End.Character != 1 {
		t.Fatalf("range = %#v", r)
	}
}

func TestLeftPad4AndStrconvItoa(t *testing.T) {
	t.Parallel()

	if got := leftPad4(7); got != "0007" {
		t.Fatalf("leftPad4(7) = %q", got)
	}
	if got := leftPad4(42); got != "0042" {
		t.Fatalf("leftPad4(42) = %q", got)
	}
	if got := leftPad4(123); got != "0123" {
		t.Fatalf("leftPad4(123) = %q", got)
	}
	if got := leftPad4(1234); got != "1234" {
		t.Fatalf("leftPad4(1234) = %q", got)
	}
	if got := strconvItoa(-9); got != "-9" {
		t.Fatalf("strconvItoa(-9) = %q", got)
	}
	if got := strconvItoa(0); got != "0" {
		t.Fatalf("strconvItoa(0) = %q", got)
	}
}

func TestKeywordSetAndExtractSymbolsNil(t *testing.T) {
	t.Parallel()

	set := keywordSet()
	if !set["if"] || !set["define"] {
		t.Fatalf("keyword set missing expected entries")
	}
	f, v := extractSymbols(nil)
	if f != nil || v != nil {
		t.Fatalf("extractSymbols(nil) = %v %v, want nil nil", f, v)
	}
}

func TestCrossDocumentDefinitionFromWorkspaceIndex(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(note(t, "textDocument/didOpen", lsproto.DidOpenTextDocumentParams{TextDocument: lsproto.TextDocumentItem{URI: "file:///lib.wrg", LanguageID: "worng", Version: 1, Text: "// call add(a,b) }\n// discard a\n// {\n"}})); err != nil {
		t.Fatalf("open lib: %v", err)
	}
	if err := s.Handle(note(t, "textDocument/didOpen", lsproto.DidOpenTextDocumentParams{TextDocument: lsproto.TextDocumentItem{URI: "file:///app.wrg", LanguageID: "worng", Version: 1, Text: "// define add(1,2)\n"}})); err != nil {
		t.Fatalf("open app: %v", err)
	}

	reqP := lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///app.wrg"}, Position: lsproto.Position{Line: 0, Character: 10}}
	if err := s.Handle(req(t, 99, "textDocument/definition", reqP)); err != nil {
		t.Fatalf("definition handle: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error != nil {
		t.Fatalf("definition error: %#v", resp.Error)
	}
	var loc lsproto.Location
	decodeResult(t, resp.Result, &loc)
	if loc.URI != "file:///lib.wrg" {
		t.Fatalf("cross-doc definition uri = %q, want file:///lib.wrg", loc.URI)
	}
}

func rawID(i int) *json.RawMessage {
	r := json.RawMessage([]byte(strconv.Itoa(i)))
	return &r
}
