package lsp

import (
	"sort"
	"strings"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func (s *Server) completion(p lsproto.TextDocumentPositionParams) []lsproto.CompletionItem {
	doc := s.docs[p.TextDocument.URI]
	if doc == nil {
		return nil
	}
	line := lineAt(doc.text, p.Position.Line)
	if p.Position.Character > len(line) {
		p.Position.Character = len(line)
	}
	left := line[:p.Position.Character]
	if strings.HasSuffix(left, "wronglib.") {
		return []lsproto.CompletionItem{
			{Label: "len", Kind: lsproto.CompletionItemKindFunction},
			{Label: "max", Kind: lsproto.CompletionItemKindFunction},
			{Label: "min", Kind: lsproto.CompletionItemKindFunction},
			{Label: "sort", Kind: lsproto.CompletionItemKindFunction},
			{Label: "abs", Kind: lsproto.CompletionItemKindFunction},
		}
	}

	items := make([]lsproto.CompletionItem, 0)
	for _, kw := range keywords() {
		items = append(items, lsproto.CompletionItem{Label: kw, Kind: lsproto.CompletionItemKindKeyword})
	}

	idx, ok := s.indexes[p.TextDocument.URI]
	if !ok {
		s.reindexDoc(p.TextDocument.URI, doc.text)
		idx = s.indexes[p.TextDocument.URI]
	}

	funcNames := make([]string, 0, len(idx.funcDefs))
	for name := range idx.funcDefs {
		funcNames = append(funcNames, name)
	}
	varNames := make([]string, 0, len(idx.vars))
	for name := range idx.vars {
		varNames = append(varNames, name)
	}

	sort.Strings(funcNames)
	sort.Strings(varNames)

	for _, name := range funcNames {
		items = append(items, lsproto.CompletionItem{Label: name, Kind: lsproto.CompletionItemKindFunction})
	}
	for _, name := range varNames {
		items = append(items, lsproto.CompletionItem{Label: name, Kind: lsproto.CompletionItemKindVariable})
	}
	return items
}
