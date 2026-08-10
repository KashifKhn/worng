package lsp

import (
	"testing"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func FuzzDocumentChanges(f *testing.F) {
	f.Add("// x = 1\n", "// x = 2", 0, 0, 0, 1, false)
	f.Add("", "replacement", 0, 0, 0, 0, true)
	f.Add("one\ntwo\n", "x", -1, 0, 99, 99, false)

	f.Fuzz(func(t *testing.T, text, replacement string, startLine, startChar, endLine, endChar int, fullReplace bool) {
		if len(text) > 64*1024 {
			text = text[:64*1024]
		}
		if len(replacement) > 16*1024 {
			replacement = replacement[:16*1024]
		}

		if fullReplace {
			changes := []lsproto.TextDocumentContentChangeEvent{{Text: replacement}}
			_ = applyIncrementalChanges(text, changes)
			return
		}

		change := lsproto.TextDocumentContentChangeEvent{
			Range: &lsproto.Range{
				Start: lsproto.Position{Line: startLine, Character: startChar},
				End:   lsproto.Position{Line: endLine, Character: endChar},
			},
			Text: replacement,
		}
		_ = applyIncrementalChanges(text, []lsproto.TextDocumentContentChangeEvent{change})
	})
}
