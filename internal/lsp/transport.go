package lsp

import (
	"encoding/json"

	"github.com/KashifKhn/worng/internal/jsonrpc"
	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func (s *Server) sendResult(id json.RawMessage, result interface{}) error {
	if result == nil {
		result = json.RawMessage("null")
	}
	return s.transmit(jsonrpc.ResponseMessage{JSONRPC: "2.0", ID: id, Result: result})
}

func (s *Server) sendError(id json.RawMessage, code int, msg string) error {
	return s.transmit(jsonrpc.ResponseMessage{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonrpc.ResponseError{Code: code, Message: msg},
	})
}

func (s *Server) publishDiagnostics(uri string, version *int, ds []lsproto.Diagnostic) error {
	params := lsproto.PublishDiagnosticsParams{URI: uri, Version: version, Diagnostics: ds}
	return s.transmit(jsonrpc.NotificationMessage{JSONRPC: "2.0", Method: "textDocument/publishDiagnostics", Params: mustRaw(params)})
}
