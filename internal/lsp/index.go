package lsp

import (
	"github.com/KashifKhn/worng/internal/ast"
	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func (s *Server) reindexDoc(uri, text string) {
	parsed := parseProgram(uri, text)
	idx := docIndex{
		funcDefs: make(map[string]lsproto.Location),
		funcMeta: make(map[string][]string),
		vars:     make(map[string]varInfo),
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
				var vloc lsproto.Location
				if loc == nil {
					vloc = lsproto.Location{URI: uri, Range: identRange(n.Pos().Line, n.Pos().Column, len(n.Name))}
				} else {
					vloc = *loc
				}
				if _, exists := idx.vars[n.Name]; !exists {
					idx.vars[n.Name] = varInfo{Location: vloc, InferredType: inferExprType(n.Value)}
				}
				idx.symbols = append(idx.symbols, lsproto.SymbolInformation{Name: n.Name, Kind: lsproto.SymbolKindVariable, Location: vloc})
			}
		}
	}

	s.indexes[uri] = idx
	s.parses[uri] = parsed
}

func inferExprType(expr ast.Expression) string {
	if expr == nil {
		return "unknown"
	}
	switch expr.(type) {
	case *ast.NumberLiteral:
		return "number"
	case *ast.StringLiteral:
		return "string"
	case *ast.BoolLiteral:
		return "bool"
	case *ast.NullLiteral:
		return "null"
	case *ast.ArrayLiteral:
		return "array"
	case *ast.FuncCallNode:
		return "function"
	case *ast.BinaryNode, *ast.UnaryNode:
		return "expression"
	default:
		return "unknown"
	}
}

func firstWordLocation(uri, text, word string) *lsproto.Location {
	locs := findWordLocations(uri, text, word)
	if len(locs) == 0 {
		return nil
	}
	return &locs[0]
}
