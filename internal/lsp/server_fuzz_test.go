package lsp

import (
	"encoding/json"
	"testing"

	"github.com/KashifKhn/worng/internal/jsonrpc"
)

func FuzzServerHandle(f *testing.F) {
	f.Add("2.0", "initialize", `{}`)
	f.Add("2.0", "textDocument/didOpen", `{"textDocument":{"uri":"file:///a","languageId":"worng","version":1,"text":"// if\n"}}`)

	f.Fuzz(func(t *testing.T, version, method, params string) {
		s := NewServer()
		id := json.RawMessage("1")
		msg := jsonrpc.Message{
			JSONRPC: version,
			ID:      &id,
			Method:  method,
			Params:  json.RawMessage(params),
		}
		_ = s.Handle(msg)
	})
}
