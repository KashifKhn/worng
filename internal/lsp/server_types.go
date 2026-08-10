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

type hoverDoc struct {
	Title   string
	Written string
	Actual  string
	Gotcha  string
	Example string
	SpecRef string
}

type docIndex struct {
	funcDefs map[string]lsproto.Location
	funcMeta map[string][]string
	vars     map[string]varInfo
	symbols  []lsproto.SymbolInformation
}

type varInfo struct {
	Location     lsproto.Location
	InferredType string
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

	keywordDoc  map[string]hoverDoc
	operatorDoc map[string]hoverDoc
	wronglibDoc map[string]hoverDoc
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
		state:       statePreInit,
		exitCode:    1,
		transmit:    func(v interface{}) error { return nil },
		debounce:    0,
		posEnc:      "utf-16",
		docs:        make(map[string]*document),
		indexes:     make(map[string]docIndex),
		parses:      make(map[string]parseResult),
		timers:      make(map[string]*time.Timer),
		canceled:    make(map[string]bool),
		keywordDoc:  defaultKeywordDocs(),
		operatorDoc: defaultOperatorDocs(),
		wronglibDoc: defaultWronglibDocs(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}
