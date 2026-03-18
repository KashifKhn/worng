package main

import (
	"encoding/json"
	"io"
	"os"

	"github.com/KashifKhn/worng/internal/jsonrpc"
	"github.com/KashifKhn/worng/internal/lsp"
)

func lspCommand() int {
	conn := jsonrpc.NewConn(os.Stdin, os.Stdout)
	srv := lsp.NewServer(lsp.WithTransmit(func(v interface{}) error {
		return conn.WriteMessage(v)
	}), lsp.WithDebounceMillis(150))

	for {
		msg, err := conn.ReadMessage()
		if err != nil {
			if err == io.EOF {
				return 0
			}
			_ = conn.WriteMessage(jsonrpc.ResponseMessage{
				JSONRPC: "2.0",
				ID:      json.RawMessage("null"),
				Error: &jsonrpc.ResponseError{
					Code:    jsonrpc.ErrorParseError,
					Message: err.Error(),
				},
			})
			continue
		}

		if err := srv.Handle(msg); err != nil {
			if msg.ID != nil {
				_ = conn.WriteMessage(jsonrpc.ResponseMessage{
					JSONRPC: "2.0",
					ID:      *msg.ID,
					Error: &jsonrpc.ResponseError{
						Code:    jsonrpc.ErrorInternalError,
						Message: err.Error(),
					},
				})
			}
		}

		if srv.Exited() {
			return srv.ExitCode()
		}
	}
}
