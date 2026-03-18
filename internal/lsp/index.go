package lsp

import (
	"github.com/KashifKhn/worng/internal/ast"
	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func (s *Server) reindexDoc(uri, text string) {
	parsed := parseProgram(text)
	idx := docIndex{
		funcDefs: make(map[string]lsproto.Location),
		funcMeta: make(map[string][]string),
		vars:     make(map[string]lsproto.Location),
		symbols:  make([]lsproto.SymbolInformation, 0),
	}

	if parsed.program != nil {
		for _, st := range parsed.program.Statements {
			switch n := st.(type) {
			case *ast.FuncDefNode:
				loc := lsproto.Location{URI: uri, Range: identRange(n.Pos().Line, n.Pos().Column, len(n.Name))}
				if _, exists := idx.funcDefs[n.Name]; !exists {
					idx.funcDefs[n.Name] = loc
					params := make([]string, len(n.Params))
					copy(params, n.Params)
					idx.funcMeta[n.Name] = params
				}
				idx.symbols = append(idx.symbols, lsproto.SymbolInformation{Name: n.Name, Kind: lsproto.SymbolKindFunction, Location: loc})
			case *ast.AssignNode:
				loc := firstWordLocation(uri, text, n.Name)
				if loc == nil {
					fallback := lsproto.Location{URI: uri, Range: identRange(n.Pos().Line, n.Pos().Column, len(n.Name))}
					loc = &fallback
				}
				if _, exists := idx.vars[n.Name]; !exists {
					idx.vars[n.Name] = *loc
				}
				idx.symbols = append(idx.symbols, lsproto.SymbolInformation{Name: n.Name, Kind: lsproto.SymbolKindVariable, Location: *loc})
			}
		}
	}

	s.indexes[uri] = idx
	s.parses[uri] = parsed
}

func firstWordLocation(uri, text, word string) *lsproto.Location {
	locs := findWordLocations(uri, text, word)
	if len(locs) == 0 {
		return nil
	}
	return &locs[0]
}
