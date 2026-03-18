// Package lsp implements the WORNG Language Server.
// It uses internal/jsonrpc for transport and internal/lsp/lsproto for protocol types.
package lsp

import (
	"encoding/json"
	"strings"

	"github.com/KashifKhn/worng/internal/jsonrpc"
	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func (s *Server) Exited() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state == stateExited
}

func (s *Server) ExitCode() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.exitCode
}

func (s *Server) Handle(msg jsonrpc.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if msg.IsRequest() {
		return s.handleRequest(msg)
	}
	if msg.IsNotification() {
		return s.handleNotification(msg)
	}
	return nil
}

func (s *Server) handleRequest(msg jsonrpc.Message) error {
	if s.state == statePreInit && msg.Method != "initialize" {
		return s.sendError(*msg.ID, jsonrpc.ErrorServerNotInit, "server not initialized")
	}
	if s.state == stateShutdown {
		return s.sendError(*msg.ID, jsonrpc.ErrorInvalidRequest, "server is shutting down")
	}
	if isCanceled(s, *msg.ID) {
		return s.sendError(*msg.ID, jsonrpc.ErrorRequestCancelled, "request cancelled")
	}

	switch msg.Method {
	case "initialize":
		if s.state != statePreInit {
			return s.sendError(*msg.ID, jsonrpc.ErrorInvalidRequest, "initialize already called")
		}
		s.posEnc = s.resolvePositionEncoding(msg.Params)
		s.state = stateInitialized
		return s.sendResult(*msg.ID, lsproto.InitializeResult{Capabilities: s.capabilities()})

	case "shutdown":
		s.state = stateShutdown
		s.exitCode = 0
		return s.sendResult(*msg.ID, nil)

	case "textDocument/hover":
		var p lsproto.TextDocumentPositionParams
		if err := json.Unmarshal(msg.Params, &p); err != nil || p.TextDocument.URI == "" {
			return s.sendError(*msg.ID, jsonrpc.ErrorInvalidParams, "invalid params")
		}
		h := s.hover(p)
		if h == nil {
			return s.sendResult(*msg.ID, nil)
		}
		return s.sendResult(*msg.ID, h)

	case "textDocument/completion":
		var p lsproto.TextDocumentPositionParams
		if err := json.Unmarshal(msg.Params, &p); err != nil || p.TextDocument.URI == "" {
			return s.sendError(*msg.ID, jsonrpc.ErrorInvalidParams, "invalid params")
		}
		return s.sendResult(*msg.ID, s.completion(p))

	case "textDocument/definition":
		var p lsproto.TextDocumentPositionParams
		if err := json.Unmarshal(msg.Params, &p); err != nil || p.TextDocument.URI == "" {
			return s.sendError(*msg.ID, jsonrpc.ErrorInvalidParams, "invalid params")
		}
		loc := s.definition(p)
		if loc == nil {
			return s.sendResult(*msg.ID, nil)
		}
		return s.sendResult(*msg.ID, loc)

	case "textDocument/references":
		var p lsproto.ReferenceParams
		if err := json.Unmarshal(msg.Params, &p); err != nil || p.TextDocument.URI == "" {
			return s.sendError(*msg.ID, jsonrpc.ErrorInvalidParams, "invalid params")
		}
		return s.sendResult(*msg.ID, s.references(p))

	case "textDocument/rename":
		var p lsproto.RenameParams
		if err := json.Unmarshal(msg.Params, &p); err != nil || p.TextDocument.URI == "" {
			return s.sendError(*msg.ID, jsonrpc.ErrorInvalidParams, "invalid params")
		}
		edit := s.rename(p)
		if edit == nil {
			return s.sendResult(*msg.ID, nil)
		}
		return s.sendResult(*msg.ID, edit)

	case "textDocument/signatureHelp":
		var p lsproto.TextDocumentPositionParams
		if err := json.Unmarshal(msg.Params, &p); err != nil || p.TextDocument.URI == "" {
			return s.sendError(*msg.ID, jsonrpc.ErrorInvalidParams, "invalid params")
		}
		h := s.signatureHelp(p)
		if h == nil {
			return s.sendResult(*msg.ID, nil)
		}
		return s.sendResult(*msg.ID, h)

	case "textDocument/formatting":
		var p lsproto.DocumentFormattingParams
		if err := json.Unmarshal(msg.Params, &p); err != nil || p.TextDocument.URI == "" {
			return s.sendError(*msg.ID, jsonrpc.ErrorInvalidParams, "invalid params")
		}
		doc := s.docs[p.TextDocument.URI]
		if doc == nil {
			return s.sendResult(*msg.ID, nil)
		}
		edits := []lsproto.TextEdit{{Range: fullDocumentRange(doc.text), NewText: formatDocument(doc.text)}}
		return s.sendResult(*msg.ID, edits)

	case "textDocument/documentSymbol":
		var p struct {
			TextDocument lsproto.TextDocumentIdentifier `json:"textDocument"`
		}
		if err := json.Unmarshal(msg.Params, &p); err != nil || p.TextDocument.URI == "" {
			return s.sendError(*msg.ID, jsonrpc.ErrorInvalidParams, "invalid params")
		}
		return s.sendResult(*msg.ID, s.documentSymbols(p.TextDocument.URI))

	case "textDocument/semanticTokens/full":
		var p struct {
			TextDocument lsproto.TextDocumentIdentifier `json:"textDocument"`
		}
		if err := json.Unmarshal(msg.Params, &p); err != nil || p.TextDocument.URI == "" {
			return s.sendError(*msg.ID, jsonrpc.ErrorInvalidParams, "invalid params")
		}
		return s.sendResult(*msg.ID, s.semanticTokens(p.TextDocument.URI))

	default:
		return s.sendError(*msg.ID, jsonrpc.ErrorMethodNotFound, "method not found")
	}
}

func (s *Server) handleNotification(msg jsonrpc.Message) error {
	if msg.Method == "exit" {
		s.state = stateExited
		if s.exitCode == 0 {
			s.exitCode = 0
		} else {
			s.exitCode = 1
		}
		return nil
	}

	if s.state == statePreInit {
		return nil
	}

	switch msg.Method {
	case "initialized":
		return nil

	case "textDocument/didOpen":
		var p lsproto.DidOpenTextDocumentParams
		if err := json.Unmarshal(msg.Params, &p); err != nil {
			return nil
		}
		s.docs[p.TextDocument.URI] = &document{uri: p.TextDocument.URI, text: p.TextDocument.Text, version: p.TextDocument.Version}
		s.reindexDoc(p.TextDocument.URI, p.TextDocument.Text)
		s.scheduleDiagnostics(p.TextDocument.URI)

	case "textDocument/didChange":
		var p lsproto.DidChangeTextDocumentParams
		if err := json.Unmarshal(msg.Params, &p); err != nil {
			return nil
		}
		doc, ok := s.docs[p.TextDocument.URI]
		if !ok || p.TextDocument.Version < doc.version {
			return nil
		}
		doc.version = p.TextDocument.Version
		if len(p.ContentChanges) > 0 {
			if p.ContentChanges[len(p.ContentChanges)-1].Range == nil {
				doc.text = p.ContentChanges[len(p.ContentChanges)-1].Text
			} else {
				doc.text = applyIncrementalChanges(doc.text, p.ContentChanges)
			}
			s.reindexDoc(p.TextDocument.URI, doc.text)
		}
		s.scheduleDiagnostics(p.TextDocument.URI)

	case "textDocument/didClose":
		var p lsproto.DidCloseTextDocumentParams
		if err := json.Unmarshal(msg.Params, &p); err != nil {
			return nil
		}
		delete(s.docs, p.TextDocument.URI)
		delete(s.indexes, p.TextDocument.URI)
		delete(s.parses, p.TextDocument.URI)
		_ = s.publishDiagnostics(p.TextDocument.URI, nil, nil)

	case "$/cancelRequest":
		var p struct {
			ID interface{} `json:"id"`
		}
		if err := json.Unmarshal(msg.Params, &p); err == nil {
			s.canceled[mustIDString(p.ID)] = true
		}
	}

	return nil
}

func (s *Server) capabilities() lsproto.ServerCapabilities {
	return lsproto.ServerCapabilities{
		PositionEncoding:           s.posEnc,
		TextDocumentSync:           lsproto.TextDocumentSyncFull,
		HoverProvider:              true,
		DefinitionProvider:         true,
		ReferencesProvider:         true,
		RenameProvider:             true,
		DocumentFormattingProvider: true,
		SignatureHelpProvider:      &lsproto.SignatureHelpOptions{TriggerCharacters: []string{"(", ","}},
		DocumentSymbolProvider:     true,
		CompletionProvider: &lsproto.CompletionOptions{
			ResolveProvider:   false,
			TriggerCharacters: []string{"."},
		},
		SemanticTokensProvider: &lsproto.SemanticTokensOptions{
			Legend: lsproto.SemanticTokensLegend{
				TokenTypes:     []string{"keyword", "variable", "function", "string", "number", "operator", "comment"},
				TokenModifiers: []string{},
			},
			Full: true,
		},
	}
}

func choosePositionEncoding(encs []string) string {
	if len(encs) == 0 {
		return "utf-16"
	}
	for _, enc := range encs {
		if strings.EqualFold(enc, "utf-8") {
			return "utf-8"
		}
		if strings.EqualFold(enc, "utf-16") {
			return "utf-16"
		}
		if strings.EqualFold(enc, "utf-32") {
			return "utf-32"
		}
	}
	return "utf-16"
}

func (s *Server) resolvePositionEncoding(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "utf-16"
	}

	var v2 lsproto.InitializeParamsV2
	if err := json.Unmarshal(raw, &v2); err == nil {
		if len(v2.Capabilities.General.PositionEncodings) > 0 {
			encs := make([]string, 0, len(v2.Capabilities.General.PositionEncodings))
			for _, v := range v2.Capabilities.General.PositionEncodings {
				encs = append(encs, string(v))
			}
			return choosePositionEncoding(encs)
		}
	}

	var legacy lsproto.InitializeParams
	if err := json.Unmarshal(raw, &legacy); err != nil {
		return "utf-16"
	}
	return choosePositionEncoding(extractPositionEncodings(legacy.Capabilities))
}

func extractPositionEncodings(caps map[string]interface{}) []string {
	if caps == nil {
		return nil
	}
	general, ok := caps["general"].(map[string]interface{})
	if !ok {
		return nil
	}
	vals, ok := general["positionEncodings"].([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(vals))
	for _, v := range vals {
		s, ok := v.(string)
		if ok {
			out = append(out, s)
		}
	}
	return out
}
