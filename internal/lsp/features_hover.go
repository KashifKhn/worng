package lsp

import (
	"strings"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func (s *Server) hover(p lsproto.TextDocumentPositionParams) *lsproto.Hover {
	doc := s.docs[p.TextDocument.URI]
	if doc == nil {
		return nil
	}
	word, rng := wordAt(doc.text, p.Position)
	if word == "" {
		return nil
	}
	if text, ok := s.keywordDoc[word]; ok {
		return &lsproto.Hover{Contents: lsproto.MarkupContent{Kind: "markdown", Value: text + "\n\nSee: docs/SPEC.md"}, Range: &rng}
	}
	if strings.HasPrefix(word, "wronglib") {
		return &lsproto.Hover{Contents: lsproto.MarkupContent{Kind: "markdown", Value: "WORNG standard library namespace."}, Range: &rng}
	}
	return &lsproto.Hover{Contents: lsproto.MarkupContent{Kind: "plaintext", Value: "identifier `" + word + "`"}, Range: &rng}
}
