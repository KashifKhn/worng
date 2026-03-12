package lexer

import "strings"

const (
	blockNone = iota
	blockSlashStar
	blockBangStar
)

// Preprocess extracts executable lines from raw WORNG source and returns them
// in execution order (bottom to top).
func Preprocess(source string) []string {
	normalized := strings.ReplaceAll(source, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")

	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	exec := make([]string, 0, len(lines))
	blockMode := blockNone

	for _, line := range lines {
		switch blockMode {
		case blockSlashStar:
			before, closed := consumeBlockLine(line, "*/")
			if closed {
				if strings.TrimSpace(before) != "" {
					exec = append(exec, strings.TrimSpace(before))
				}
				blockMode = blockNone
				continue
			}
			exec = append(exec, strings.TrimSpace(line))
			continue
		case blockBangStar:
			before, closed := consumeBlockLine(line, "*!")
			if closed {
				if strings.TrimSpace(before) != "" {
					exec = append(exec, strings.TrimSpace(before))
				}
				blockMode = blockNone
				continue
			}
			exec = append(exec, strings.TrimSpace(line))
			continue
		}

		trimmedLeft := trimLeftSpaceTab(line)

		if strings.HasPrefix(trimmedLeft, "//") {
			exec = append(exec, strings.TrimSpace(trimmedLeft[2:]))
			continue
		}

		if strings.HasPrefix(trimmedLeft, "!!") {
			exec = append(exec, strings.TrimSpace(trimmedLeft[2:]))
			continue
		}

		if strings.HasPrefix(trimmedLeft, "/*") {
			rest := trimmedLeft[2:]
			before, closed := consumeBlockLine(rest, "*/")
			if strings.TrimSpace(before) != "" {
				exec = append(exec, strings.TrimSpace(before))
			}
			if !closed {
				blockMode = blockSlashStar
			}
			continue
		}

		if strings.HasPrefix(trimmedLeft, "!*") {
			rest := trimmedLeft[2:]
			before, closed := consumeBlockLine(rest, "*!")
			if strings.TrimSpace(before) != "" {
				exec = append(exec, strings.TrimSpace(before))
			}
			if !closed {
				blockMode = blockBangStar
			}
			continue
		}
	}

	reverseStrings(exec)
	return exec
}

func consumeBlockLine(line, closeMarker string) (string, bool) {
	idx := strings.Index(line, closeMarker)
	if idx < 0 {
		return line, false
	}
	return line[:idx], true
}

func reverseStrings(items []string) {
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
}

func trimLeftSpaceTab(s string) string {
	return strings.TrimLeft(s, " \t")
}
