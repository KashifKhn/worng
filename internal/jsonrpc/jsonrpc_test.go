package jsonrpc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestReadMessageValidRequest(t *testing.T) {
	t.Parallel()

	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"x":1}}`
	framed := "Content-Length: " + strconv.Itoa(len(body)) + "\r\n\r\n" + body

	msg, err := ReadOneMessage(strings.NewReader(framed))
	if err != nil {
		t.Fatalf("ReadOneMessage error: %v", err)
	}
	if !msg.IsRequest() {
		t.Fatalf("message is not request: %#v", msg)
	}
	if msg.Method != "initialize" {
		t.Fatalf("method = %q, want initialize", msg.Method)
	}
}

func TestReadMessageRejectsInvalidJSONRPCVersion(t *testing.T) {
	t.Parallel()

	body := `{"jsonrpc":"1.0","id":1,"method":"initialize"}`
	framed := "Content-Length: " + strconv.Itoa(len(body)) + "\r\n\r\n" + body

	_, err := ReadOneMessage(strings.NewReader(framed))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReadMessageMissingContentLength(t *testing.T) {
	t.Parallel()

	_, err := ReadOneMessage(strings.NewReader("Content-Type: application/vscode-jsonrpc; charset=utf-8\r\n\r\n{}"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "content-length") {
		t.Fatalf("error = %v, want content-length mention", err)
	}
}

func TestReadMessageInvalidHeaderLine(t *testing.T) {
	t.Parallel()

	_, err := ReadOneMessage(strings.NewReader("NoColonHere\r\n\r\n"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReadMessageInvalidContentLength(t *testing.T) {
	t.Parallel()

	_, err := ReadOneMessage(strings.NewReader("Content-Length: abc\r\n\r\n"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReadMessageRejectsUnsupportedCharset(t *testing.T) {
	t.Parallel()

	_, err := ReadOneMessage(strings.NewReader("Content-Length: 2\r\nContent-Type: application/vscode-jsonrpc; charset=latin1\r\n\r\n{}"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "charset") {
		t.Fatalf("error = %v, want charset mention", err)
	}
}

func TestReadMessageAcceptsUtf8Alias(t *testing.T) {
	t.Parallel()

	body := `{"jsonrpc":"2.0","method":"initialized"}`
	framed := "Content-Length: " + strconv.Itoa(len(body)) + "\r\nContent-Type: application/vscode-jsonrpc; charset=utf8\r\n\r\n" + body

	msg, err := ReadOneMessage(strings.NewReader(framed))
	if err != nil {
		t.Fatalf("ReadOneMessage error: %v", err)
	}
	if !msg.IsNotification() {
		t.Fatalf("message is not notification: %#v", msg)
	}
}

func TestWriteAndReadRoundTrip(t *testing.T) {
	t.Parallel()

	input := ResponseMessage{
		JSONRPC: jsonRPCVersion,
		ID:      json.RawMessage("1"),
		Result:  map[string]string{"ok": "yes"},
	}

	frame, err := EncodeToFrame(input)
	if err != nil {
		t.Fatalf("EncodeToFrame error: %v", err)
	}

	msg, err := ReadOneMessage(bytes.NewReader(frame))
	if err != nil {
		t.Fatalf("ReadOneMessage error: %v", err)
	}
	if !msg.IsResponse() {
		t.Fatalf("message is not response: %#v", msg)
	}
}

func TestReadMessageInvalidJSONPayload(t *testing.T) {
	t.Parallel()

	body := `{this-is-not-json}`
	framed := "Content-Length: " + strconv.Itoa(len(body)) + "\r\n\r\n" + body

	_, err := ReadOneMessage(strings.NewReader(framed))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReadMessageShortPayload(t *testing.T) {
	t.Parallel()

	body := `{"jsonrpc":"2.0"}`
	framed := "Content-Length: " + strconv.Itoa(len(body)+10) + "\r\n\r\n" + body

	_, err := ReadOneMessage(strings.NewReader(framed))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteMessageMarshalError(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	conn := NewConn(strings.NewReader(""), &out)
	err := conn.WriteMessage(func() {})
	if err == nil {
		t.Fatal("expected marshal error, got nil")
	}
}

func TestWriteMessageHeaderWriteError(t *testing.T) {
	t.Parallel()

	conn := NewConn(strings.NewReader(""), &failingWriter{failOnWrite: 1})
	err := conn.WriteMessage(map[string]string{"ok": "yes"})
	if err == nil {
		t.Fatal("expected write error, got nil")
	}
}

func TestWriteMessageBodyWriteError(t *testing.T) {
	t.Parallel()

	conn := NewConn(strings.NewReader(""), &failingWriter{failOnWrite: 2})
	err := conn.WriteMessage(map[string]string{"ok": "yes"})
	if err == nil {
		t.Fatal("expected write error, got nil")
	}
}

func TestValidateContentTypeNoCharsetAndEmpty(t *testing.T) {
	t.Parallel()

	if err := validateContentType(""); err != nil {
		t.Fatalf("validate empty error: %v", err)
	}
	if err := validateContentType("application/vscode-jsonrpc"); err != nil {
		t.Fatalf("validate no-charset error: %v", err)
	}
	if err := validateContentType("application/vscode-jsonrpc; charset=utf-8; q=1"); err != nil {
		t.Fatalf("validate charset with suffix error: %v", err)
	}
}

func TestEncodeToFrameMarshalError(t *testing.T) {
	t.Parallel()

	_, err := EncodeToFrame(make(chan int))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestWriteErrorSetsNullIDWhenMissing(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	conn := NewConn(strings.NewReader(""), &out)
	if err := conn.WriteError(nil, ErrorInvalidRequest, "bad", nil); err != nil {
		t.Fatalf("WriteError error: %v", err)
	}

	msg, err := ReadOneMessage(bytes.NewReader(out.Bytes()))
	if err != nil {
		t.Fatalf("ReadOneMessage error: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte(`"id":null`)) {
		t.Fatalf("wire response missing null id: %q", out.String())
	}
	if msg.Error == nil || msg.Error.Code != ErrorInvalidRequest {
		t.Fatalf("error = %#v, want code %d", msg.Error, ErrorInvalidRequest)
	}
}

func TestReadHeadersRejectsOversizedHeaders(t *testing.T) {
	t.Parallel()

	large := strings.Repeat("A", maxHeaderBytes+10)
	_, err := readHeaders(bufReader(large + "\r\n"))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestConnConcurrentWritesAreFramed(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	conn := NewConn(strings.NewReader(""), &out)

	const n = 20
	var wg sync.WaitGroup
	wg.Add(n)
	errCh := make(chan error, n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			errCh <- conn.WriteResponse(json.RawMessage("1"), map[string]int{"i": i})
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent write error: %v", err)
		}
	}

	if got := strings.Count(out.String(), "Content-Length: "); got != n {
		t.Fatalf("frame count = %d, want %d", got, n)
	}

	readConn := NewConn(bytes.NewReader(out.Bytes()), io.Discard)
	parsed := 0
	for {
		_, err := readConn.ReadMessage()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("parse frame %d failed: %v", parsed, err)
		}
		parsed++
	}
	if parsed != n {
		t.Fatalf("parsed frames = %d, want %d", parsed, n)
	}
}

func TestMessageClassification(t *testing.T) {
	t.Parallel()

	id := json.RawMessage("1")
	req := Message{JSONRPC: jsonRPCVersion, Method: "m", ID: &id}
	if !req.IsRequest() || req.IsNotification() || req.IsResponse() {
		t.Fatalf("request classification wrong: %#v", req)
	}

	not := Message{JSONRPC: jsonRPCVersion, Method: "m"}
	if !not.IsNotification() || not.IsRequest() || not.IsResponse() {
		t.Fatalf("notification classification wrong: %#v", not)
	}

	resp := Message{JSONRPC: jsonRPCVersion, ID: &id}
	if !resp.IsResponse() || resp.IsRequest() || resp.IsNotification() {
		t.Fatalf("response classification wrong: %#v", resp)
	}
}

func FuzzReadOneMessage(f *testing.F) {
	f.Add("Content-Length: 2\r\n\r\n{}")
	f.Add("Content-Length: 40\r\n\r\n{\"jsonrpc\":\"2.0\",\"method\":\"x\"}")

	f.Fuzz(func(t *testing.T, input string) {
		_, _ = ReadOneMessage(strings.NewReader(input))
	})
}

func bufReader(s string) *bufio.Reader {
	return bufio.NewReader(strings.NewReader(s))
}

type failingWriter struct {
	failOnWrite int
	writes      int
}

func (w *failingWriter) Write(p []byte) (int, error) {
	w.writes++
	if w.writes == w.failOnWrite {
		return 0, io.ErrClosedPipe
	}
	return len(p), nil
}
