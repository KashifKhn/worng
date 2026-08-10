package lsp

import (
	"fmt"
	"strings"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func applyIncrementalChanges(text string, changes []lsproto.TextDocumentContentChangeEvent) string {
	out := text
	for _, c := range changes {
		if c.Range == nil {
			out = c.Text
			continue
		}
		start := positionToOffset(out, c.Range.Start)
		end := positionToOffset(out, c.Range.End)
		if start < 0 || end < start || end > len(out) {
			continue
		}
		out = out[:start] + c.Text + out[end:]
	}
	return out
}

func positionToOffset(text string, p lsproto.Position) int {
	if p.Line < 0 || p.Character < 0 {
		return -1
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if p.Line >= len(lines) {
		return -1
	}
	off := 0
	for i := 0; i < p.Line; i++ {
		off += len(lines[i]) + 1
	}
	if p.Character > len(lines[p.Line]) {
		return -1
	}
	return off + p.Character
}

func mustIDString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}
