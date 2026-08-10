package jsonrpc

import (
	"bytes"
	"testing"
)

func FuzzJSONRPCTransport(f *testing.F) {
	f.Add([]byte("Content-Length: 31\r\n\r\n{\"jsonrpc\":\"2.0\",\"id\":1}"))
	f.Add([]byte("Content-Length: 0\r\n\r\n"))
	f.Add([]byte("not a JSON-RPC frame"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 64*1024 {
			data = data[:64*1024]
		}
		_, _ = ReadOneMessage(bytes.NewReader(data))
	})
}
