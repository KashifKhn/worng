package lsp

import (
	"strings"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func findWordLocations(uri, text, word string) []lsproto.Location {
	if word == "" {
		return nil
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	out := make([]lsproto.Location, 0)
	for li, line := range lines {
		start := 0
		for {
			idx := strings.Index(line[start:], word)
			if idx < 0 {
				break
			}
			at := start + idx
			leftOK := at == 0 || !isWordPart(rune(line[at-1]))
			rightEnd := at + len(word)
			rightOK := rightEnd == len(line) || !isWordPart(rune(line[rightEnd]))
			if leftOK && rightOK {
				out = append(out, lsproto.Location{
					URI: uri,
					Range: lsproto.Range{
						Start: lsproto.Position{Line: li, Character: at},
						End:   lsproto.Position{Line: li, Character: rightEnd},
					},
				})
			}
			start = at + len(word)
			if start >= len(line) {
				break
			}
		}
	}
	return out
}
