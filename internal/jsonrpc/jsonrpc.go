// Package jsonrpc implements JSON-RPC 2.0 transport framing.
package jsonrpc

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
)

const (
	headerContentLength = "content-length"
	headerContentType   = "content-type"
	maxHeaderBytes      = 16 * 1024
)

type Conn struct {
	r  *bufio.Reader
	w  io.Writer
	mu sync.Mutex
}

func NewConn(r io.Reader, w io.Writer) *Conn {
	return &Conn{r: bufio.NewReader(r), w: w}
}

func (c *Conn) ReadMessage() (Message, error) {
	contentLength, err := readHeaders(c.r)
	if err != nil {
		return Message{}, err
	}

	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(c.r, payload); err != nil {
		return Message{}, err
	}

	var msg Message
	if err := json.Unmarshal(payload, &msg); err != nil {
		return Message{}, err
	}
	if msg.JSONRPC != jsonRPCVersion {
		return Message{}, fmt.Errorf("unsupported jsonrpc version %q", msg.JSONRPC)
	}
	return msg, nil
}

func (c *Conn) WriteMessage(v interface{}) error {
	payload, err := json.Marshal(v)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(payload))

	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := io.WriteString(c.w, header); err != nil {
		return err
	}
	_, err = c.w.Write(payload)
	return err
}

func (c *Conn) WriteResponse(id json.RawMessage, result interface{}) error {
	return c.WriteMessage(ResponseMessage{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Result:  result,
	})
}

func (c *Conn) WriteError(id json.RawMessage, code int, message string, data interface{}) error {
	if len(id) == 0 {
		id = json.RawMessage("null")
	}
	return c.WriteMessage(ResponseMessage{
		JSONRPC: jsonRPCVersion,
		ID:      id,
		Error: &ResponseError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	})
}

func ReadOneMessage(r io.Reader) (Message, error) {
	conn := NewConn(r, io.Discard)
	return conn.ReadMessage()
}

func WriteOneMessage(w io.Writer, v interface{}) error {
	conn := NewConn(strings.NewReader(""), w)
	return conn.WriteMessage(v)
}

func readHeaders(r *bufio.Reader) (int, error) {
	contentLength := -1
	seenBytes := 0

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return 0, err
		}
		seenBytes += len(line)
		if seenBytes > maxHeaderBytes {
			return 0, errors.New("jsonrpc header too large")
		}

		if line == "\r\n" {
			break
		}

		key, value, ok := splitHeaderLine(line)
		if !ok {
			return 0, fmt.Errorf("invalid header line %q", strings.TrimSpace(line))
		}

		switch strings.ToLower(key) {
		case headerContentLength:
			n, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || n < 0 {
				return 0, fmt.Errorf("invalid content-length %q", strings.TrimSpace(value))
			}
			contentLength = n
		case headerContentType:
			if err := validateContentType(value); err != nil {
				return 0, err
			}
		}
	}

	if contentLength < 0 {
		return 0, errors.New("missing content-length header")
	}
	return contentLength, nil
}

func splitHeaderLine(line string) (key, value string, ok bool) {
	idx := strings.Index(line, ":")
	if idx <= 0 {
		return "", "", false
	}
	key = strings.TrimSpace(line[:idx])
	value = strings.TrimSpace(line[idx+1:])
	value = strings.TrimSuffix(value, "\r")
	return key, value, true
}

func validateContentType(ct string) error {
	if ct == "" {
		return nil
	}
	lower := strings.ToLower(strings.TrimSpace(ct))
	if !strings.Contains(lower, "charset=") {
		return nil
	}
	idx := strings.Index(lower, "charset=")
	charset := strings.TrimSpace(lower[idx+len("charset="):])
	if semi := strings.Index(charset, ";"); semi >= 0 {
		charset = strings.TrimSpace(charset[:semi])
	}
	if charset != "utf-8" && charset != "utf8" {
		return fmt.Errorf("unsupported charset %q", charset)
	}
	return nil
}

func EncodeToFrame(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := WriteOneMessage(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
