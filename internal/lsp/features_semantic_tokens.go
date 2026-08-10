package lsp

import (
	"sort"
	"strings"
	"unicode"

	"github.com/KashifKhn/worng/internal/lsp/lsproto"
)

func (s *Server) semanticTokens(uri string) lsproto.SemanticTokens {
	doc := s.docs[uri]
	if doc == nil {
		return lsproto.SemanticTokens{}
	}

	lines := strings.Split(strings.ReplaceAll(doc.text, "\r\n", "\n"), "\n")
	tokens := make([][5]int, 0)
	kw := keywordSet()

	for li, line := range lines {
		trim := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trim, "//") || strings.HasPrefix(trim, "!!") {
			ci := strings.Index(line, trim[:2])
			tokens = append(tokens, [5]int{li, ci, 2, 6, 0})
		}

		for wi := 0; wi < len(line); {
			r := rune(line[wi])
			if isWordStart(r) {
				start := wi
				wi++
				for wi < len(line) && isWordPart(rune(line[wi])) {
					wi++
				}
				word := line[start:wi]
				typeIdx := 1
				if kw[word] {
					typeIdx = 0
				}
				tokens = append(tokens, [5]int{li, start, len(word), typeIdx, 0})
				continue
			}
			if unicode.IsDigit(r) {
				start := wi
				wi++
				for wi < len(line) && (unicode.IsDigit(rune(line[wi])) || line[wi] == '.') {
					wi++
				}
				tokens = append(tokens, [5]int{li, start, wi - start, 4, 0})
				continue
			}
			if strings.ContainsRune("+-*/%=<>!", r) {
				tokens = append(tokens, [5]int{li, wi, 1, 5, 0})
			}
			wi++
		}
	}

	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i][0] == tokens[j][0] {
			return tokens[i][1] < tokens[j][1]
		}
		return tokens[i][0] < tokens[j][0]
	})

	data := make([]int, 0, len(tokens)*5)
	prevLine := 0
	prevChar := 0
	for i, tok := range tokens {
		deltaLine := tok[0] - prevLine
		deltaChar := tok[1]
		if i > 0 && deltaLine == 0 {
			deltaChar = tok[1] - prevChar
		}
		data = append(data, deltaLine, deltaChar, tok[2], tok[3], tok[4])
		prevLine = tok[0]
		prevChar = tok[1]
	}

	return lsproto.SemanticTokens{Data: data}
}
