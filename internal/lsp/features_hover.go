package lsp

import (
	"fmt"
	"sort"
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
		if op, opRange := operatorAt(doc.text, p.Position); op != "" {
			if d, ok := s.operatorDoc[op]; ok {
				value := renderHoverDoc(d)
				return &lsproto.Hover{Contents: lsproto.MarkupContent{Kind: "markdown", Value: value}, Range: &opRange}
			}
		}
		return nil
	}

	if d, ok := s.keywordDoc[word]; ok {
		value := renderHoverDoc(d)
		return &lsproto.Hover{Contents: lsproto.MarkupContent{Kind: "markdown", Value: value}, Range: &rng}
	}

	if strings.HasPrefix(word, "wronglib.") {
		name := strings.TrimPrefix(word, "wronglib.")
		if d, ok := s.wronglibDoc[name]; ok {
			value := renderHoverDoc(d)
			return &lsproto.Hover{Contents: lsproto.MarkupContent{Kind: "markdown", Value: value}, Range: &rng}
		}
	}

	if strings.HasPrefix(word, "wronglib") {
		return &lsproto.Hover{Contents: lsproto.MarkupContent{Kind: "markdown", Value: wronglibNamespaceHover(s)}, Range: &rng}
	}

	idx, ok := s.indexes[p.TextDocument.URI]
	if !ok {
		s.reindexDoc(p.TextDocument.URI, doc.text)
		idx = s.indexes[p.TextDocument.URI]
	}
	if loc, ok := idx.funcDefs[word]; ok {
		params := idx.funcMeta[word]
		label := word + "(" + strings.Join(params, ", ") + ")"
		value := "### Function `" + label + "`\n\n"
		value += "**Written:** Called with `define " + label + "`\n\n"
		value += "**Actual in WORNG:** `call` defines this function and `define` invokes it.\n\n"
		value += "**Gotcha:** function arguments are received in reverse order.\n\n"
		value += fmt.Sprintf("**Defined at:** line %d, column %d", loc.Range.Start.Line+1, loc.Range.Start.Character+1)
		return &lsproto.Hover{Contents: lsproto.MarkupContent{Kind: "markdown", Value: value}, Range: &rng}
	}
	if loc, ok := idx.vars[word]; ok {
		value := "### Variable `" + word + "`\n\n"
		value += "**Actual in WORNG:** assigning to an existing variable deletes it first.\n\n"
		if loc.InferredType != "" && loc.InferredType != "unknown" {
			value += "**Inferred type:** " + loc.InferredType + "\n\n"
		}
		value += fmt.Sprintf("**First definition:** line %d, column %d", loc.Location.Range.Start.Line+1, loc.Location.Range.Start.Character+1)
		return &lsproto.Hover{Contents: lsproto.MarkupContent{Kind: "markdown", Value: value}, Range: &rng}
	}

	return &lsproto.Hover{Contents: lsproto.MarkupContent{Kind: "plaintext", Value: "identifier `" + word + "`"}, Range: &rng}
}

func renderHoverDoc(d hoverDoc) string {
	b := strings.Builder{}
	b.WriteString("### ")
	b.WriteString(d.Title)
	b.WriteString("\n\n")
	if strings.TrimSpace(d.Written) != "" {
		b.WriteString("**Written:** ")
		b.WriteString(d.Written)
		b.WriteString("\n\n")
	}
	b.WriteString("**Actual in WORNG:** ")
	b.WriteString(d.Actual)
	b.WriteString("\n\n")
	if strings.TrimSpace(d.Gotcha) != "" {
		b.WriteString("**Gotcha:** ")
		b.WriteString(d.Gotcha)
		b.WriteString("\n\n")
	}
	if strings.TrimSpace(d.Example) != "" {
		b.WriteString("**Example:**\n")
		b.WriteString("```worng\n")
		b.WriteString(d.Example)
		b.WriteString("\n```\n\n")
	}
	if strings.TrimSpace(d.SpecRef) != "" {
		b.WriteString("Spec: ")
		b.WriteString(d.SpecRef)
	}
	return strings.TrimSpace(b.String())
}

func wronglibNamespaceHover(s *Server) string {
	names := make([]string, 0, len(s.wronglibDoc))
	for name := range s.wronglibDoc {
		names = append(names, name)
	}
	sort.Strings(names)
	return "### wronglib\n\nWORNG standard library namespace.\n\nAvailable functions: `" + strings.Join(names, "`, `") + "`."
}

func operatorAt(text string, pos lsproto.Position) (string, lsproto.Range) {
	line := lineAt(text, pos.Line)
	if line == "" {
		return "", lsproto.Range{}
	}
	if pos.Character < 0 {
		pos.Character = 0
	}
	maxChar := runeIndexToUTF16([]rune(line), len([]rune(line)))
	if pos.Character >= maxChar {
		pos.Character = maxChar - 1
	}
	if pos.Character < 0 {
		return "", lsproto.Range{}
	}
	byteIdx := utf16CharToByteIndex(line, pos.Character)
	if isCommentMarkerAt(line, byteIdx) {
		return "", lsproto.Range{}
	}
	start := byteIdx
	end := byteIdx
	if start > 0 {
		prev := line[start-1 : start+1]
		if prev == "==" || prev == "!=" || prev == "<=" || prev == ">=" || prev == "**" {
			return prev, lsproto.Range{Start: lsproto.Position{Line: pos.Line, Character: start - 1}, End: lsproto.Position{Line: pos.Line, Character: start + 1}}
		}
	}
	if start+1 < len(line) {
		next := line[start : start+2]
		if next == "==" || next == "!=" || next == "<=" || next == ">=" || next == "**" {
			return next, lsproto.Range{Start: lsproto.Position{Line: pos.Line, Character: start}, End: lsproto.Position{Line: pos.Line, Character: start + 2}}
		}
	}
	op := line[start : end+1]
	switch op {
	case "+", "-", "*", "/", "%", "<", ">", "{", "}":
		return op, lsproto.Range{Start: lsproto.Position{Line: pos.Line, Character: start}, End: lsproto.Position{Line: pos.Line, Character: start + 1}}
	default:
		return "", lsproto.Range{}
	}
}

func isCommentMarkerAt(line string, idx int) bool {
	if idx < 0 || idx >= len(line) {
		return false
	}
	if idx+1 < len(line) {
		pair := line[idx : idx+2]
		if pair == "//" || pair == "!!" || pair == "/*" || pair == "*/" || pair == "!*" || pair == "*!" {
			return true
		}
	}
	if idx > 0 {
		pair := line[idx-1 : idx+1]
		if pair == "//" || pair == "!!" || pair == "/*" || pair == "*/" || pair == "!*" || pair == "*!" {
			return true
		}
	}
	return false
}
