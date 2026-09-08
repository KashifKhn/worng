package main

import (
	"strings"

	"github.com/KashifKhn/worng/internal/lexer"
)

// prepareSource runs the preprocessor and returns the prepared text plus the
// prepared-line → source-line map used to keep diagnostics pointing at real
// file locations.
func prepareSource(source string) (string, []int, error) {
	lines, err := lexer.PreprocessWithLines(source)
	if err != nil {
		return "", nil, err
	}
	lineMap := make([]int, len(lines))
	for idx, l := range lines {
		lineMap[idx] = l.SourceLine
	}
	return joinExecutableLineSlice(lines), lineMap, nil
}

func joinExecutableLineSlice(lines []lexer.ExecLine) string {
	if len(lines) == 0 {
		return ""
	}
	parts := make([]string, len(lines))
	for idx, l := range lines {
		parts[idx] = l.Content
	}
	return strings.Join(parts, "\n") + "\n"
}
