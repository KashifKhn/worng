package lsp

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/KashifKhn/worng/internal/jsonrpc"
	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func TestInitializeLifecycle(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := NewServer(WithTransmit(tx.send), WithDebounceMillis(0))

	if err := s.Handle(req(t, 1, "initialize", lsproto.InitializeParams{Capabilities: map[string]interface{}{}})); err != nil {
		t.Fatalf("initialize error: %v", err)
	}
	if err := s.Handle(note(t, "initialized", map[string]interface{}{})); err != nil {
		t.Fatalf("initialized error: %v", err)
	}

	resp := tx.lastResponse(t)
	if resp.Error != nil {
		t.Fatalf("initialize response error: %#v", resp.Error)
	}
	var res lsproto.InitializeResult
	decodeResult(t, resp.Result, &res)
	if res.Capabilities.TextDocumentSync != lsproto.TextDocumentSyncFull {
		t.Fatalf("sync kind = %d, want %d", res.Capabilities.TextDocumentSync, lsproto.TextDocumentSyncFull)
	}
	if !res.Capabilities.HoverProvider || !res.Capabilities.DefinitionProvider || !res.Capabilities.DocumentSymbolProvider {
		t.Fatalf("missing mandatory caps: %#v", res.Capabilities)
	}
	if !res.Capabilities.ReferencesProvider || !res.Capabilities.RenameProvider || !res.Capabilities.DocumentFormattingProvider {
		t.Fatalf("missing advanced capabilities: %#v", res.Capabilities)
	}
	if res.Capabilities.SignatureHelpProvider == nil {
		t.Fatal("signature help provider missing")
	}
	if res.Capabilities.CompletionProvider == nil {
		t.Fatal("completion provider missing")
	}
	if res.Capabilities.SemanticTokensProvider == nil {
		t.Fatal("semantic tokens provider missing")
	}
}

func TestRequestBeforeInitialize(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := NewServer(WithTransmit(tx.send), WithDebounceMillis(0))
	if err := s.Handle(req(t, 2, "textDocument/hover", lsproto.TextDocumentPositionParams{})); err != nil {
		t.Fatalf("handle error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error == nil || resp.Error.Code != jsonrpc.ErrorServerNotInit {
		t.Fatalf("response error = %#v, want code %d", resp.Error, jsonrpc.ErrorServerNotInit)
	}
}

func TestInitializeTwiceReturnsInvalidRequest(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := NewServer(WithTransmit(tx.send), WithDebounceMillis(0))
	if err := s.Handle(req(t, 1, "initialize", lsproto.InitializeParams{Capabilities: map[string]interface{}{}})); err != nil {
		t.Fatalf("initialize error: %v", err)
	}
	if err := s.Handle(req(t, 2, "initialize", lsproto.InitializeParams{Capabilities: map[string]interface{}{}})); err != nil {
		t.Fatalf("second initialize handle error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error == nil || resp.Error.Code != jsonrpc.ErrorInvalidRequest {
		t.Fatalf("response error = %#v, want code %d", resp.Error, jsonrpc.ErrorInvalidRequest)
	}
}

func TestShutdownThenExit(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(req(t, 4, "shutdown", nil)); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
	if err := s.Handle(note(t, "exit", nil)); err != nil {
		t.Fatalf("exit error: %v", err)
	}
	if !s.Exited() {
		t.Fatal("server should be exited")
	}
	if s.ExitCode() != 0 {
		t.Fatalf("exit code = %d, want 0", s.ExitCode())
	}
}

func TestExitWithoutShutdownHasErrorExitCode(t *testing.T) {
	t.Parallel()

	s := NewServer(WithDebounceMillis(0))
	if err := s.Handle(note(t, "exit", nil)); err != nil {
		t.Fatalf("exit error: %v", err)
	}
	if s.ExitCode() != 1 {
		t.Fatalf("exit code = %d, want 1", s.ExitCode())
	}
}

func TestDidOpenPublishesSyntaxDiagnostic(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)

	open := lsproto.DidOpenTextDocumentParams{TextDocument: lsproto.TextDocumentItem{URI: "file:///a.wrg", LanguageID: "worng", Version: 1, Text: "// if\n"}}
	if err := s.Handle(note(t, "textDocument/didOpen", open)); err != nil {
		t.Fatalf("didOpen error: %v", err)
	}

	pub := tx.lastNotification(t, "textDocument/publishDiagnostics")
	var p lsproto.PublishDiagnosticsParams
	decodeParams(t, pub.Params, &p)
	if len(p.Diagnostics) == 0 {
		t.Fatal("expected diagnostics, got none")
	}
	if p.Diagnostics[0].Severity != lsproto.DiagnosticSeverityError {
		t.Fatalf("severity = %d, want %d", p.Diagnostics[0].Severity, lsproto.DiagnosticSeverityError)
	}
}

func TestDidChangeClearsDiagnosticsOnFix(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(note(t, "textDocument/didOpen", lsproto.DidOpenTextDocumentParams{TextDocument: lsproto.TextDocumentItem{URI: "file:///a.wrg", LanguageID: "worng", Version: 1, Text: "// if\n"}})); err != nil {
		t.Fatalf("didOpen error: %v", err)
	}
	change := lsproto.DidChangeTextDocumentParams{
		TextDocument:   lsproto.VersionedTextDocumentIdentifier{URI: "file:///a.wrg", Version: 2},
		ContentChanges: []lsproto.TextDocumentContentChangeEvent{{Text: "// input ~\"ok\"\n"}},
	}
	if err := s.Handle(note(t, "textDocument/didChange", change)); err != nil {
		t.Fatalf("didChange error: %v", err)
	}
	pub := tx.lastNotification(t, "textDocument/publishDiagnostics")
	var p lsproto.PublishDiagnosticsParams
	decodeParams(t, pub.Params, &p)
	if len(p.Diagnostics) != 0 {
		t.Fatalf("expected no diagnostics, got %d", len(p.Diagnostics))
	}
}

func TestDidChangeStaleVersionIgnored(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(note(t, "textDocument/didOpen", lsproto.DidOpenTextDocumentParams{TextDocument: lsproto.TextDocumentItem{URI: "file:///a.wrg", LanguageID: "worng", Version: 3, Text: "// input ~\"ok\"\n"}})); err != nil {
		t.Fatalf("didOpen error: %v", err)
	}
	stale := lsproto.DidChangeTextDocumentParams{
		TextDocument:   lsproto.VersionedTextDocumentIdentifier{URI: "file:///a.wrg", Version: 2},
		ContentChanges: []lsproto.TextDocumentContentChangeEvent{{Text: "// if\n"}},
	}
	if err := s.Handle(note(t, "textDocument/didChange", stale)); err != nil {
		t.Fatalf("didChange stale error: %v", err)
	}
	pub := tx.lastNotification(t, "textDocument/publishDiagnostics")
	var p lsproto.PublishDiagnosticsParams
	decodeParams(t, pub.Params, &p)
	if len(p.Diagnostics) != 0 {
		t.Fatalf("stale change should not reintroduce diagnostics; got %d", len(p.Diagnostics))
	}
}

func TestDidClosePublishesEmptyDiagnostics(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///a.wrg", "// if\n")
	if err := s.Handle(note(t, "textDocument/didClose", lsproto.DidCloseTextDocumentParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a.wrg"}})); err != nil {
		t.Fatalf("didClose error: %v", err)
	}
	pub := tx.lastNotification(t, "textDocument/publishDiagnostics")
	var p lsproto.PublishDiagnosticsParams
	decodeParams(t, pub.Params, &p)
	if len(p.Diagnostics) != 0 {
		t.Fatalf("expected zero diagnostics on close, got %d", len(p.Diagnostics))
	}
}

func TestHoverKeyword(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///h.wrg", "// if false }\n// input ~\"x\"\n// {\n")
	reqP := lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///h.wrg"}, Position: lsproto.Position{Line: 0, Character: 3}}
	if err := s.Handle(req(t, 5, "textDocument/hover", reqP)); err != nil {
		t.Fatalf("hover handle error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error != nil {
		t.Fatalf("hover response error: %#v", resp.Error)
	}
	var h lsproto.Hover
	decodeResult(t, resp.Result, &h)
	if !strings.Contains(strings.ToLower(h.Contents.Value), "false") {
		t.Fatalf("hover value = %q, want inversion content", h.Contents.Value)
	}
}

func TestHoverNoWordReturnsNull(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///h.wrg", "// input ~\"x\"\n")
	reqP := lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///h.wrg"}, Position: lsproto.Position{Line: 0, Character: 0}}
	if err := s.Handle(req(t, 6, "textDocument/hover", reqP)); err != nil {
		t.Fatalf("hover handle error: %v", err)
	}
	resp := tx.lastResponse(t)
	if strings.TrimSpace(string(resp.Result)) != "null" {
		t.Fatalf("hover result = %s, want null", string(resp.Result))
	}
}

func TestCompletionKeywordsAndWronglib(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///c.wrg", "// define wronglib.\n")
	reqP := lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///c.wrg"}, Position: lsproto.Position{Line: 0, Character: 17}}
	if err := s.Handle(req(t, 7, "textDocument/completion", reqP)); err != nil {
		t.Fatalf("completion handle error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error != nil {
		t.Fatalf("completion response error: %#v", resp.Error)
	}
	var items []lsproto.CompletionItem
	decodeResult(t, resp.Result, &items)
	if len(items) < 5 {
		t.Fatalf("completion items = %d, want >= 5", len(items))
	}
}

func TestDefinitionFunction(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///d.wrg", "// call add(a, b) }\n// discard a\n// {\n// define add(1, 2)\n")
	reqP := lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///d.wrg"}, Position: lsproto.Position{Line: 3, Character: 11}}
	if err := s.Handle(req(t, 8, "textDocument/definition", reqP)); err != nil {
		t.Fatalf("definition handle error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error != nil {
		t.Fatalf("definition response error: %#v", resp.Error)
	}
	var loc lsproto.Location
	decodeResult(t, resp.Result, &loc)
	if loc.URI != "file:///d.wrg" {
		t.Fatalf("definition uri = %q, want file:///d.wrg", loc.URI)
	}
}

func TestDefinitionNotFoundReturnsNull(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///d.wrg", "// input ~\"x\"\n")
	reqP := lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///d.wrg"}, Position: lsproto.Position{Line: 0, Character: 3}}
	if err := s.Handle(req(t, 9, "textDocument/definition", reqP)); err != nil {
		t.Fatalf("definition handle error: %v", err)
	}
	resp := tx.lastResponse(t)
	if strings.TrimSpace(string(resp.Result)) != "null" {
		t.Fatalf("definition result = %s, want null", string(resp.Result))
	}
}

func TestDocumentSymbols(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///s.wrg", "// x = 1\n// call f(a) }\n// discard a\n// {\n")
	if err := s.Handle(req(t, 10, "textDocument/documentSymbol", map[string]interface{}{"textDocument": lsproto.TextDocumentIdentifier{URI: "file:///s.wrg"}})); err != nil {
		t.Fatalf("documentSymbol handle error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error != nil {
		t.Fatalf("documentSymbol response error: %#v", resp.Error)
	}
	var syms []lsproto.SymbolInformation
	decodeResult(t, resp.Result, &syms)
	if len(syms) < 2 {
		t.Fatalf("symbols = %d, want >= 2", len(syms))
	}
}

func TestSemanticTokensFull(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initializedWithDoc(t, tx, "file:///t.wrg", "// x = 1\n// input x\n")
	if err := s.Handle(req(t, 11, "textDocument/semanticTokens/full", map[string]interface{}{"textDocument": lsproto.TextDocumentIdentifier{URI: "file:///t.wrg"}})); err != nil {
		t.Fatalf("semanticTokens handle error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error != nil {
		t.Fatalf("semanticTokens response error: %#v", resp.Error)
	}
	var st lsproto.SemanticTokens
	decodeResult(t, resp.Result, &st)
	if len(st.Data) == 0 {
		t.Fatal("semantic tokens are empty")
	}
}

func TestUnknownMethodForRequestReturnsMethodNotFound(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(req(t, 12, "no/suchMethod", nil)); err != nil {
		t.Fatalf("unknown method handle error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error == nil || resp.Error.Code != jsonrpc.ErrorMethodNotFound {
		t.Fatalf("error = %#v, want code %d", resp.Error, jsonrpc.ErrorMethodNotFound)
	}
}

func TestShutdownBlocksFurtherRequests(t *testing.T) {
	t.Parallel()

	tx := &capture{}
	s := initialized(t, tx)
	if err := s.Handle(req(t, 13, "shutdown", nil)); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
	if err := s.Handle(req(t, 14, "textDocument/hover", lsproto.TextDocumentPositionParams{})); err != nil {
		t.Fatalf("post-shutdown request error: %v", err)
	}
	resp := tx.lastResponse(t)
	if resp.Error == nil || resp.Error.Code != jsonrpc.ErrorInvalidRequest {
		t.Fatalf("post-shutdown error = %#v, want code %d", resp.Error, jsonrpc.ErrorInvalidRequest)
	}
}

type capture struct {
	msgs []jsonrpc.Message
}

func (c *capture) send(v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	var m jsonrpc.Message
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	c.msgs = append(c.msgs, m)
	return nil
}

func (c *capture) lastResponse(tb testing.TB) jsonrpc.Message {
	tb.Helper()
	for i := len(c.msgs) - 1; i >= 0; i-- {
		if c.msgs[i].IsResponse() {
			return c.msgs[i]
		}
	}
	tb.Fatal("no response recorded")
	return jsonrpc.Message{}
}

func (c *capture) lastNotification(tb testing.TB, method string) jsonrpc.Message {
	tb.Helper()
	for i := len(c.msgs) - 1; i >= 0; i-- {
		if c.msgs[i].IsNotification() && c.msgs[i].Method == method {
			return c.msgs[i]
		}
	}
	tb.Fatalf("notification %q not found", method)
	return jsonrpc.Message{}
}

func initialized(tb testing.TB, c *capture) *Server {
	tb.Helper()
	s := NewServer(WithTransmit(c.send), WithDebounceMillis(0))
	if err := s.Handle(req(tb, 1, "initialize", lsproto.InitializeParams{Capabilities: map[string]interface{}{}})); err != nil {
		tb.Fatalf("initialize failed: %v", err)
	}
	if err := s.Handle(note(tb, "initialized", map[string]interface{}{})); err != nil {
		tb.Fatalf("initialized failed: %v", err)
	}
	return s
}

func initializedWithDoc(tb testing.TB, c *capture, uri, text string) *Server {
	tb.Helper()
	s := initialized(tb, c)
	open := lsproto.DidOpenTextDocumentParams{TextDocument: lsproto.TextDocumentItem{URI: uri, LanguageID: "worng", Version: 1, Text: text}}
	if err := s.Handle(note(tb, "textDocument/didOpen", open)); err != nil {
		tb.Fatalf("didOpen failed: %v", err)
	}
	return s
}

func req(tb testing.TB, id int, method string, params interface{}) jsonrpc.Message {
	tb.Helper()
	pb, err := json.Marshal(params)
	if err != nil {
		tb.Fatalf("marshal params: %v", err)
	}
	rawID := json.RawMessage(strconv.AppendInt(nil, int64(id), 10))
	return jsonrpc.Message{JSONRPC: "2.0", ID: &rawID, Method: method, Params: pb}
}

func note(tb testing.TB, method string, params interface{}) jsonrpc.Message {
	tb.Helper()
	pb, err := json.Marshal(params)
	if err != nil {
		tb.Fatalf("marshal params: %v", err)
	}
	return jsonrpc.Message{JSONRPC: "2.0", Method: method, Params: pb}
}

func decodeParams(tb testing.TB, raw json.RawMessage, out interface{}) {
	tb.Helper()
	if err := json.Unmarshal(raw, out); err != nil {
		tb.Fatalf("unmarshal params: %v", err)
	}
}

func decodeResult(tb testing.TB, raw json.RawMessage, out interface{}) {
	tb.Helper()
	if err := json.Unmarshal(raw, out); err != nil {
		tb.Fatalf("unmarshal result: %v", err)
	}
}
