package lsp

import (
	"strings"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func (s *Server) references(p lsproto.ReferenceParams) []lsproto.Location {
	doc := s.docs[p.TextDocument.URI]
	if doc == nil {
		return nil
	}
	word, _ := wordAt(doc.text, p.Position)
	if word == "" {
		return nil
	}

	out := make([]lsproto.Location, 0)
	for uri, d := range s.docs {
		if len(s.canceled) > 0 {
			return out
		}
		out = append(out, findWordLocations(uri, d.text, word)...)
	}

	if !p.Context.IncludeDeclaration {
		decl := s.lookupDeclaration(word)
		if decl != nil {
			filtered := make([]lsproto.Location, 0, len(out))
			for _, loc := range out {
				if !sameLocation(loc, *decl) {
					filtered = append(filtered, loc)
				}
			}
			out = filtered
		}
	}

	return out
}

func (s *Server) rename(p lsproto.RenameParams) *lsproto.WorkspaceEdit {
	if strings.TrimSpace(p.NewName) == "" {
		return nil
	}
	doc := s.docs[p.TextDocument.URI]
	if doc == nil {
		return nil
	}
	word, _ := wordAt(doc.text, p.Position)
	if word == "" {
		return nil
	}

	changes := make(map[string][]lsproto.TextEdit)
	for uri, d := range s.docs {
		if len(s.canceled) > 0 {
			return nil
		}
		locs := findWordLocations(uri, d.text, word)
		edits := make([]lsproto.TextEdit, 0, len(locs))
		for _, loc := range locs {
			edits = append(edits, lsproto.TextEdit{Range: loc.Range, NewText: p.NewName})
		}
		if len(edits) > 0 {
			changes[uri] = edits
		}
	}

	if len(changes) == 0 {
		return nil
	}
	return &lsproto.WorkspaceEdit{Changes: changes}
}

func (s *Server) lookupDeclaration(word string) *lsproto.Location {
	for _, idx := range s.indexes {
		if loc, ok := idx.funcDefs[word]; ok {
			return &loc
		}
		if loc, ok := idx.vars[word]; ok {
			return &loc
		}
	}
	return nil
}

func sameLocation(a, b lsproto.Location) bool {
	return a.URI == b.URI &&
		a.Range.Start.Line == b.Range.Start.Line &&
		a.Range.Start.Character == b.Range.Start.Character &&
		a.Range.End.Line == b.Range.End.Line &&
		a.Range.End.Character == b.Range.End.Character
}
