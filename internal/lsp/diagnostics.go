package lsp

import (
	"time"

	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func (s *Server) scheduleDiagnostics(uri string) {
	if s.debounce <= 0 {
		s.publishFromDoc(uri)
		return
	}
	if tm, ok := s.timers[uri]; ok {
		tm.Stop()
	}
	s.timers[uri] = time.AfterFunc(s.debounce, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.publishFromDoc(uri)
	})
}

func (s *Server) publishFromDoc(uri string) {
	doc, ok := s.docs[uri]
	if !ok {
		return
	}
	parsed := parseProgram(doc.text)
	s.parses[uri] = parsed
	items := make([]lsproto.Diagnostic, 0, len(parsed.errs))
	for _, err := range parsed.errs {
		we, ok := err.(*diagnostics.WorngError)
		if !ok {
			continue
		}
		line := we.Pos.Line - 1
		if line < 0 {
			line = 0
		}
		char := we.Pos.Column - 1
		if char < 0 {
			char = 0
		}
		items = append(items, lsproto.Diagnostic{
			Range: lsproto.Range{
				Start: lsproto.Position{Line: line, Character: 0},
				End:   lsproto.Position{Line: line, Character: 1024},
			},
			Severity: lsproto.DiagnosticSeverityError,
			Code:     "W" + leftPad4(we.Diag.Code),
			Source:   "worng",
			Message:  we.Error(),
		})
	}
	_ = s.publishDiagnostics(uri, &doc.version, items)
}
