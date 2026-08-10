package lsp

import (
	"strings"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func formatDocument(text string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	out := strings.Join(lines, "\n")
	out = strings.TrimRight(out, "\n")
	if out != "" {
		out += "\n"
	}
	return out
}

func fullDocumentRange(text string) lsproto.Range {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 0 {
		return lsproto.Range{Start: lsproto.Position{}, End: lsproto.Position{}}
	}
	last := lines[len(lines)-1]
	if len(lines) > 1 && last == "" {
		last = lines[len(lines)-2]
	}
	endLine := len(lines) - 1
	if len(lines) > 1 && lines[len(lines)-1] == "" {
		endLine = len(lines) - 2
	}
	if endLine < 0 {
		endLine = 0
	}
	return lsproto.Range{
		Start: lsproto.Position{Line: 0, Character: 0},
		End:   lsproto.Position{Line: endLine, Character: len(last)},
	}
}
