package jsonrpc

import "encoding/json"

const jsonRPCVersion = "2.0"

const (
	ErrorParseError         = -32700
	ErrorInvalidRequest     = -32600
	ErrorMethodNotFound     = -32601
	ErrorInvalidParams      = -32602
	ErrorInternalError      = -32603
	ErrorServerNotInit      = -32002
	ErrorUnknownServerError = -32001
	ErrorRequestCancelled   = -32800
)

type RequestMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type NotificationMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type ResponseError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ResponseMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *ResponseError  `json:"error,omitempty"`
}

type Message struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method,omitempty"`
	Params  json.RawMessage  `json:"params,omitempty"`
	Result  json.RawMessage  `json:"result,omitempty"`
	Error   *ResponseError   `json:"error,omitempty"`
}

func (m Message) IsRequest() bool {
	return m.Method != "" && m.ID != nil
}

func (m Message) IsNotification() bool {
	return m.Method != "" && m.ID == nil
}

func (m Message) IsResponse() bool {
	return m.Method == "" && m.ID != nil
}
