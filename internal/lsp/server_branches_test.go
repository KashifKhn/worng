package lsp

import (
	"testing"
	"time"

	"github.com/KashifKhn/worng/internal/ast"
	"github.com/KashifKhn/worng/internal/jsonrpc"
	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func TestRequestPreInitShutdownRejectedAsNotInitialized(t *testing.T) {
	t.Parallel()

	c := &capture{}
	s := NewServer(WithTransmit(c.send))
	if err := s.Handle(req(t, 20, "shutdown", nil)); err != nil {
		t.Fatalf("handle shutdown error: %v", err)
	}
	resp := c.lastResponse(t)
	if resp.Error == nil || resp.Error.Code != jsonrpc.ErrorServerNotInit {
		t.Fatalf("error = %#v, want %d", resp.Error, jsonrpc.ErrorServerNotInit)
	}
}

func TestPublishFromDocNoDiagnostics(t *testing.T) {
	t.Parallel()

	c := &capture{}
	s := initialized(t, c)
	s.docs["file:///clean.wrg"] = &document{uri: "file:///clean.wrg", text: "// input ~\"ok\"\n", version: 2}
	s.publishFromDoc("file:///clean.wrg")

	msg := c.lastNotification(t, "textDocument/publishDiagnostics")
	var p lsproto.PublishDiagnosticsParams
	decodeParams(t, msg.Params, &p)
	if len(p.Diagnostics) != 0 {
		t.Fatalf("diagnostics = %d, want 0", len(p.Diagnostics))
	}
}

func TestCompletionBranches(t *testing.T) {
	t.Parallel()

	s := NewServer()
	if got := s.completion(lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///none"}}); got != nil {
		t.Fatalf("completion for missing doc = %#v, want nil", got)
	}

	s.docs["file:///kw"] = &document{uri: "file:///kw", text: "// whi\n", version: 1}
	items := s.completion(lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///kw"}, Position: lsproto.Position{Line: 0, Character: 4}})
	if len(items) == 0 {
		t.Fatal("expected keyword completions")
	}

	s.docs["file:///syms"] = &document{uri: "file:///syms", text: "// x = 1\n// call add(a) }\n// discard a\n// {\n", version: 1}
	items = s.completion(lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///syms"}, Position: lsproto.Position{Line: 0, Character: 999}})
	if len(items) == 0 {
		t.Fatal("expected symbol-aware completions")
	}
}

func TestJoinLinesAndLineAtBranches(t *testing.T) {
	t.Parallel()

	if got := joinLines(nil); got != "" {
		t.Fatalf("joinLines(nil) = %q, want empty", got)
	}
	text := "a\nb"
	if got := lineAt(text, -1); got != "" {
		t.Fatalf("lineAt(-1) = %q, want empty", got)
	}
	if got := lineAt(text, 5); got != "" {
		t.Fatalf("lineAt(5) = %q, want empty", got)
	}
}

func TestNewServerDefaults(t *testing.T) {
	t.Parallel()

	s := NewServer()
	if s.state != statePreInit {
		t.Fatalf("state = %v, want preInit", s.state)
	}
	if s.exitCode != 1 {
		t.Fatalf("exitCode = %d, want 1", s.exitCode)
	}
	if s.transmit == nil {
		t.Fatal("transmit should not be nil")
	}
	if len(s.keywordDoc) == 0 {
		t.Fatal("keyword docs should be populated")
	}
}

func TestScheduleDiagnosticsDebouncedAndReplaceTimer(t *testing.T) {
	t.Parallel()

	c := &capture{}
	s := initialized(t, c)
	s.debounce = 2 * time.Millisecond
	s.docs["file:///deb.wrg"] = &document{uri: "file:///deb.wrg", text: "// if\n", version: 1}

	s.scheduleDiagnostics("file:///deb.wrg")
	s.scheduleDiagnostics("file:///deb.wrg")
	time.Sleep(10 * time.Millisecond)

	_ = c.lastNotification(t, "textDocument/publishDiagnostics")
}

func TestPublishFromDocMissingDocumentNoop(t *testing.T) {
	t.Parallel()

	s := NewServer()
	s.publishFromDoc("file:///none.wrg")
}

func TestHoverBranches(t *testing.T) {
	t.Parallel()

	s := NewServer()
	if got := s.hover(lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///none"}}); got != nil {
		t.Fatalf("hover for missing doc = %#v, want nil", got)
	}

	s.docs["file:///a"] = &document{uri: "file:///a", text: "// wronglib\n", version: 1}
	h := s.hover(lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///a"}, Position: lsproto.Position{Line: 0, Character: 3}})
	if h == nil {
		t.Fatal("hover for identifier should not be nil")
	}

	s.docs["file:///b"] = &document{uri: "file:///b", text: "// wronglib.\n", version: 1}
	h2 := s.hover(lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///b"}, Position: lsproto.Position{Line: 0, Character: 5}})
	if h2 == nil {
		t.Fatal("hover for wronglib should not be nil")
	}
}

func TestDefinitionAssignBranchAndMissingDoc(t *testing.T) {
	t.Parallel()

	s := NewServer()
	if loc := s.definition(lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///none"}}); loc != nil {
		t.Fatalf("definition missing doc = %#v, want nil", loc)
	}

	s.docs["file:///d"] = &document{uri: "file:///d", text: "// x = 1\n// input x\n", version: 1}
	loc := s.definition(lsproto.TextDocumentPositionParams{TextDocument: lsproto.TextDocumentIdentifier{URI: "file:///d"}, Position: lsproto.Position{Line: 1, Character: 9}})
	if loc == nil {
		t.Fatal("expected definition location for assignment")
	}
}

func TestDocumentSymbolsAndSemanticTokensNoDoc(t *testing.T) {
	t.Parallel()

	s := NewServer()
	if syms := s.documentSymbols("file:///none"); syms != nil {
		t.Fatalf("symbols missing doc = %#v, want nil", syms)
	}
	st := s.semanticTokens("file:///none")
	if len(st.Data) != 0 {
		t.Fatalf("semantic tokens missing doc = %#v, want empty", st)
	}
}

func TestDocumentSymbolsReindexPath(t *testing.T) {
	t.Parallel()

	s := NewServer()
	s.docs["file:///sym.wrg"] = &document{uri: "file:///sym.wrg", text: "// call f(a) }\n// discard a\n// {\n", version: 1}
	syms := s.documentSymbols("file:///sym.wrg")
	if len(syms) == 0 {
		t.Fatal("expected symbols after on-demand reindex")
	}
}

func TestWordAtOutOfBoundsLine(t *testing.T) {
	t.Parallel()

	word, _ := wordAt("x", lsproto.Position{Line: 10, Character: 0})
	if word != "" {
		t.Fatalf("word = %q, want empty", word)
	}
}

func TestExtractSymbolsDedup(t *testing.T) {
	t.Parallel()

	program := &ast.ProgramNode{Statements: []ast.Statement{
		&ast.FuncDefNode{Name: "f"},
		&ast.FuncDefNode{Name: "f"},
		&ast.AssignNode{Name: "x"},
		&ast.AssignNode{Name: "x"},
	}}

	funcs, vars := extractSymbols(program)
	if len(funcs) != 1 || funcs[0] != "f" {
		t.Fatalf("funcs = %#v, want [f]", funcs)
	}
	if len(vars) != 1 || vars[0] != "x" {
		t.Fatalf("vars = %#v, want [x]", vars)
	}
}
