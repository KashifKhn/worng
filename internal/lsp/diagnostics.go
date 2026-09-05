package lsp

import (
	"fmt"
	"strings"
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
	parsed := parseProgram(uri, doc.text)
	s.parses[uri] = parsed

	allErrs := append([]error{}, parsed.errs...)
	allErrs = append(allErrs, semanticDiagnostics(parsed.program, uri)...)

	items := make([]lsproto.Diagnostic, 0, len(allErrs))
	for _, err := range allErrs {
		we, ok := err.(*diagnostics.WorngError)
		if !ok {
			continue
		}
		rng := diagRange(we.Pos)
		msg := diagnosticMessage(we)
		code := "W" + leftPad4(we.Diag.Code)
		data := diagnosticData(we)
		items = append(items, lsproto.Diagnostic{
			Range:           rng,
			Severity:        lsproto.DiagnosticSeverityError,
			Code:            code,
			CodeDescription: &lsproto.CodeDescription{Href: specRefURL(we.Diag.Key)},
			Source:          "worng",
			Message:         msg,
			Data:            data,
		})
	}
	_ = s.publishDiagnostics(uri, &doc.version, items)
}

func diagRange(pos diagnostics.Position) lsproto.Range {
	startLine := pos.Line - 1
	if startLine < 0 {
		startLine = 0
	}
	startChar := pos.Column - 1
	if startChar < 0 {
		startChar = 0
	}

	endLine := pos.EndLine - 1
	if endLine < 0 {
		endLine = startLine
	}
	endChar := pos.EndColumn
	if endChar <= startChar {
		endChar = startChar + 1
	}
	if endLine < startLine {
		endLine = startLine
	}
	if endLine == startLine && endChar <= startChar {
		endChar = startChar + 1
	}

	return lsproto.Range{
		Start: lsproto.Position{Line: startLine, Character: startChar},
		End:   lsproto.Position{Line: endLine, Character: endChar},
	}
}

func diagnosticMessage(we *diagnostics.WorngError) string {
	msg := we.Message()
	if strings.TrimSpace(we.Detail) != "" {
		msg += "\n" + we.Detail
	}
	if strings.TrimSpace(we.Hint) != "" {
		msg += "\nHint: " + we.Hint
	}
	return msg
}

func diagnosticData(we *diagnostics.WorngError) map[string]interface{} {
	data := map[string]interface{}{
		"key": we.Diag.Key,
	}
	if strings.TrimSpace(we.Hint) != "" {
		data["hint"] = we.Hint
	}
	if strings.TrimSpace(we.Detail) != "" {
		data["detail"] = we.Detail
	}
	if len(we.Expected) > 0 {
		data["expected"] = append([]string(nil), we.Expected...)
	}
	if strings.TrimSpace(we.Found) != "" {
		data["found"] = we.Found
	}
	return data
}

func specRefURL(key string) string {
	base := "https://github.com/KashifKhn/worng/blob/main/docs/SPEC.md"
	if strings.TrimSpace(key) == "" {
		return base
	}
	anchor := specrefAnchor(key)
	if anchor == "" {
		return base
	}
	return fmt.Sprintf("%s#%s", base, anchor)
}

func specrefAnchor(key string) string {
	if strings.TrimSpace(key) == "" {
		return ""
	}
	switch key {
	case "syntax_error":
		return "execution-model"
	case "undefined_variable":
		return "execution-model"
	case "type_mismatch":
		return "numbers"
	case "division_by_zero":
		return "numbers"
	case "stack_overflow":
		return "execution-model"
	case "index_out_of_bounds":
		return "execution-model"
	case "module_not_found":
		return "execution-model"
	case "file_not_found":
		return "execution-model"
	case "infinite_loop":
		return "inversion-rules"
	case "arity_mismatch":
		return "functions"
	case "invalid_number":
		return "numbers"
	default:
		return "execution-model"
	}
}
