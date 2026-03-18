package lsp

import (
	"strings"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func (s *Server) signatureHelp(p lsproto.TextDocumentPositionParams) *lsproto.SignatureHelp {
	doc := s.docs[p.TextDocument.URI]
	if doc == nil {
		return nil
	}
	line := lineAt(doc.text, p.Position.Line)
	if line == "" {
		return nil
	}
	if p.Position.Character > len(line) {
		p.Position.Character = len(line)
	}
	left := line[:p.Position.Character]
	idx := strings.LastIndex(left, "define ")
	if idx < 0 {
		return nil
	}
	call := left[idx+len("define "):]
	open := strings.Index(call, "(")
	if open <= 0 {
		return nil
	}
	name := strings.TrimSpace(call[:open])
	params := s.functionParams(name)
	if len(params) == 0 {
		return nil
	}
	commaCount := strings.Count(call[open+1:], ",")
	sig := lsproto.SignatureInformation{
		Label:      name + "(" + strings.Join(params, ", ") + ")",
		Parameters: make([]lsproto.ParameterInformation, 0, len(params)),
	}
	for _, p := range params {
		sig.Parameters = append(sig.Parameters, lsproto.ParameterInformation{Label: p})
	}
	if commaCount >= len(params) {
		commaCount = len(params) - 1
	}
	if commaCount < 0 {
		commaCount = 0
	}
	return &lsproto.SignatureHelp{Signatures: []lsproto.SignatureInformation{sig}, ActiveSignature: 0, ActiveParameter: commaCount}
}

func (s *Server) functionParams(name string) []string {
	for _, idx := range s.indexes {
		params, ok := idx.funcMeta[name]
		if ok {
			cp := make([]string, len(params))
			copy(cp, params)
			return cp
		}
	}
	for uri, d := range s.docs {
		s.reindexDoc(uri, d.text)
		params, ok := s.indexes[uri].funcMeta[name]
		if ok {
			cp := make([]string, len(params))
			copy(cp, params)
			return cp
		}
	}
	return nil
}
