package lexer

import (
	"strings"

	"github.com/KashifKhn/worng/internal/diagnostics"
)

const (
	blockNone = iota
	blockSlashStar
	blockBangStar
)

// ExecLine is one extracted executable line together with its 1-based line
// number in the original source file.
type ExecLine struct {
	Content    string
	SourceLine int
}

// Preprocess extracts executable lines from raw WORNG source in source order.
// It returns a W1012 UnterminatedBlockComment error when a block comment is
// never closed before end of file; the lines extracted so far are still
// returned so callers can render partial output alongside the diagnostic.
func Preprocess(source string) ([]string, error) {
	lines, err := PreprocessWithLines(source)
	out := make([]string, 0, len(lines))
	for _, l := range lines {
		out = append(out, l.Content)
	}
	return out, err
}

// PreprocessWithLines is Preprocess, but each returned line carries the
// 1-based line number it occupied in the original source. Callers that feed
// extracted lines to the lexer can use the map to report diagnostics against
// real file positions.
func PreprocessWithLines(source string) ([]ExecLine, error) {
	normalized := strings.ReplaceAll(source, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	exec := make([]ExecLine, 0, len(lines))
	blockMode := blockNone
	openLine := 0
	openCol := 1

	for lineIdx, line := range lines {
		srcLine := lineIdx + 1
		switch blockMode {
		case blockSlashStar:
			before, closed := consumeBlockLine(line, "*/")
			if closed {
				exec = appendBlockContent(exec, before, srcLine)
				blockMode = blockNone
				continue
			}
			// Per SPEC §4.5 block comments do not nest: an inner opener
			// marker is inert content, never a new block opener. Emitting
			// it would make the lexer re-open a comment that can never be
			// closed, so it is dropped instead of becoming code.
			if isBlockOpenerLine(line) {
				continue
			}
			exec = append(exec, ExecLine{Content: strings.TrimSpace(line), SourceLine: srcLine})
			continue
		case blockBangStar:
			before, closed := consumeBlockLine(line, "*!")
			if closed {
				exec = appendBlockContent(exec, before, srcLine)
				blockMode = blockNone
				continue
			}
			if isBlockOpenerLine(line) {
				continue
			}
			exec = append(exec, ExecLine{Content: strings.TrimSpace(line), SourceLine: srcLine})
			continue
		}

		trimmedLeft := trimLeftSpaceTab(line)

		if strings.HasPrefix(trimmedLeft, "//") {
			exec = append(exec, ExecLine{Content: strings.TrimSpace(trimmedLeft[2:]), SourceLine: srcLine})
			continue
		}

		if strings.HasPrefix(trimmedLeft, "!!") {
			exec = append(exec, ExecLine{Content: strings.TrimSpace(trimmedLeft[2:]), SourceLine: srcLine})
			continue
		}

		if strings.HasPrefix(trimmedLeft, "/*") {
			rest := trimmedLeft[2:]
			before, closed := consumeBlockLine(rest, "*/")
			exec = appendBlockContent(exec, before, srcLine)
			if !closed {
				blockMode = blockSlashStar
				openLine = srcLine
				openCol = len(line) - len(trimmedLeft) + 1
			}
			continue
		}

		if strings.HasPrefix(trimmedLeft, "!*") {
			rest := trimmedLeft[2:]
			before, closed := consumeBlockLine(rest, "*!")
			exec = appendBlockContent(exec, before, srcLine)
			if !closed {
				blockMode = blockBangStar
				openLine = srcLine
				openCol = len(line) - len(trimmedLeft) + 1
			}
			continue
		}
	}

	if blockMode != blockNone {
		var opener string
		if blockMode == blockSlashStar {
			opener = "/*"
		} else {
			opener = "!*"
		}
		return exec, diagnostics.NewUnterminatedBlockComment(
			diagnostics.Position{Line: openLine, Column: openCol, EndLine: openLine, EndColumn: openCol + 1},
			opener,
		)
	}

	return exec, nil
}

func consumeBlockLine(line, closeMarker string) (string, bool) {
	idx := strings.Index(line, closeMarker)
	if idx < 0 {
		return line, false
	}
	return line[:idx], true
}

func appendBlockContent(exec []ExecLine, content string, srcLine int) []ExecLine {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" || isBlockOpenerLine(trimmed) {
		return exec
	}
	return append(exec, ExecLine{Content: trimmed, SourceLine: srcLine})
}

func isBlockOpenerLine(s string) bool {
	t := trimLeftSpaceTab(s)
	return strings.HasPrefix(t, "/*") || strings.HasPrefix(t, "!*")
}

func trimLeftSpaceTab(s string) string {
	return strings.TrimLeft(s, " \t")
}
