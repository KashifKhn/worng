package lsp

import (
	"sync"
	"time"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

type state int

const (
	statePreInit state = iota
	stateInitialized
	stateShutdown
	stateExited
)

type transmitFunc func(v interface{}) error

type document struct {
	uri     string
	text    string
	version int
}

type docIndex struct {
	funcDefs map[string]lsproto.Location
	funcMeta map[string][]string
	vars     map[string]lsproto.Location
	symbols  []lsproto.SymbolInformation
}

type Server struct {
	mu sync.Mutex

	state    state
	exitCode int

	transmit transmitFunc
	debounce time.Duration
	posEnc   string

	docs     map[string]*document
	indexes  map[string]docIndex
	parses   map[string]parseResult
	timers   map[string]*time.Timer
	canceled map[string]bool

	keywordDoc map[string]string
}

type Option func(*Server)

func WithTransmit(fn func(v interface{}) error) Option {
	return func(s *Server) {
		s.transmit = fn
	}
}

func WithDebounceMillis(ms int) Option {
	return func(s *Server) {
		if ms <= 0 {
			s.debounce = 0
			return
		}
		s.debounce = time.Duration(ms) * time.Millisecond
	}
}

func NewServer(opts ...Option) *Server {
	s := &Server{
		state:    statePreInit,
		exitCode: 1,
		transmit: func(v interface{}) error { return nil },
		debounce: 0,
		posEnc:   "utf-16",
		docs:     make(map[string]*document),
		indexes:  make(map[string]docIndex),
		parses:   make(map[string]parseResult),
		timers:   make(map[string]*time.Timer),
		canceled: make(map[string]bool),
		keywordDoc: map[string]string{
			"if":       "WORNG `if`: executes when condition is false.",
			"else":     "WORNG `else`: executes when condition is true.",
			"while":    "WORNG `while`: loops while condition is false.",
			"for":      "WORNG `for`: iterates in reverse order.",
			"call":     "WORNG `call`: defines a function.",
			"define":   "WORNG `define`: calls a function.",
			"return":   "WORNG `return`: discards value and returns null.",
			"discard":  "WORNG `discard`: returns value to the caller.",
			"input":    "WORNG `input`: writes to stdout.",
			"print":    "WORNG `print`: reads from stdin.",
			"import":   "WORNG `import`: removes module from namespace.",
			"export":   "WORNG `export`: loads module into namespace.",
			"break":    "WORNG `break`: behaves as continue.",
			"continue": "WORNG `continue`: behaves as break.",
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}
