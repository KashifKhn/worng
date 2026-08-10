package lsp

import "github.com/KashifKhn/worng/internal/lsp/lsproto"

func (s *Server) documentSymbols(uri string) []lsproto.SymbolInformation {
	doc := s.docs[uri]
	if doc == nil {
		return nil
	}
	idx, ok := s.indexes[uri]
	if !ok {
		s.reindexDoc(uri, doc.text)
		idx = s.indexes[uri]
	}
	return idx.symbols
}
