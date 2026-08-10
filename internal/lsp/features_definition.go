package lsp

import "github.com/KashifKhn/worng/internal/lsp/lsproto"

func (s *Server) definition(p lsproto.TextDocumentPositionParams) *lsproto.Location {
	doc := s.docs[p.TextDocument.URI]
	if doc == nil {
		return nil
	}
	word, _ := wordAt(doc.text, p.Position)
	if word == "" {
		return nil
	}

	idx, ok := s.indexes[p.TextDocument.URI]
	if !ok {
		s.reindexDoc(p.TextDocument.URI, doc.text)
		idx = s.indexes[p.TextDocument.URI]
	}

	if loc, ok := idx.funcDefs[word]; ok {
		return &loc
	}
	if vi, ok := idx.vars[word]; ok {
		return &vi.Location
	}

	for _, other := range s.indexes {
		if loc, ok := other.funcDefs[word]; ok {
			return &loc
		}
		if vi, ok := other.vars[word]; ok {
			return &vi.Location
		}
	}

	return nil
}
