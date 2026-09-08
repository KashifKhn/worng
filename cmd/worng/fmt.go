package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/KashifKhn/worng/internal/lexer"
	"github.com/KashifKhn/worng/internal/vfs"
)

func fmtCommand(args []string) int {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: worng fmt <file>")
		return 2
	}
	if err := formatFile(vfs.OsFS{}, args[0]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

// formatFile normalizes a WORNG source file in place. The formatter is
// semantics-preserving: executable lines keep the markers (//, !!, /* */,
// !* *!) that make them executable — only their content is trimmed — and
// decorative plain-text lines pass through untouched. Stripping markers
// would silently turn a runnable program into plain text.
func formatFile(fs vfs.FS, path string) error {
	data, err := fs.ReadFile(path)
	if err != nil {
		return err
	}
	out, err := formatSource(string(data))
	if err != nil {
		return err
	}
	return fs.WriteFile(path, []byte(out))
}

const (
	fmtBlockNone = iota
	fmtBlockSlashStar
	fmtBlockBangStar
)

// normalizeMarkedLine trims the content of a line-comment executable line
// ("//" or "!!") while keeping its marker. A marker with no content stays
// bare ("//"), not "// " with a trailing space.
func normalizeMarkedLine(line, marker string) string {
	content := strings.TrimSpace(line[len(marker):])
	if content == "" {
		return marker
	}
	return marker + " " + content
}

// formatSource normalizes one raw WORNG file into an equivalent one.
func formatSource(source string) (string, error) {
	// Validate first: an unterminated block comment is a W1012 diagnostic,
	// and reformatting broken source would silently drop the error.
	if _, err := lexer.Preprocess(source); err != nil {
		return "", err
	}

	normalized := strings.ReplaceAll(source, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	out := make([]string, 0, len(lines))
	mode := fmtBlockNone

	for _, line := range lines {
		switch mode {
		case fmtBlockSlashStar, fmtBlockBangStar:
			closeMarker := "*/"
			if mode == fmtBlockBangStar {
				closeMarker = "*!"
			}
			if idx := strings.Index(line, closeMarker); idx >= 0 {
				// Content before the closer is executable; keep it bare.
				if before := strings.TrimSpace(line[:idx]); before != "" {
					out = append(out, before)
				}
				out = append(out, closeMarker)
				mode = fmtBlockNone
				continue
			}
			out = append(out, strings.TrimSpace(line))
			continue
		}

		trimmedLeft := strings.TrimLeft(line, " \t")
		switch {
		case strings.HasPrefix(trimmedLeft, "//"):
			out = append(out, normalizeMarkedLine(trimmedLeft, "//"))
		case strings.HasPrefix(trimmedLeft, "!!"):
			out = append(out, normalizeMarkedLine(trimmedLeft, "!!"))
		case strings.HasPrefix(trimmedLeft, "/*"), strings.HasPrefix(trimmedLeft, "!*"):
			opener := trimmedLeft[:2]
			rest := strings.TrimSpace(trimmedLeft[2:])
			closeMarker := "*/"
			if opener == "!*" {
				closeMarker = "*!"
			}
			if idx := strings.Index(rest, closeMarker); idx >= 0 {
				// Single-line block: keep opener and closer together.
				content := strings.TrimSpace(rest[:idx])
				if content == "" {
					out = append(out, opener+" "+closeMarker)
				} else {
					out = append(out, opener+" "+content+" "+closeMarker)
				}
				continue
			}
			if rest == "" {
				out = append(out, opener)
			} else {
				out = append(out, opener+" "+rest)
			}
			if opener == "/*" {
				mode = fmtBlockSlashStar
			} else {
				mode = fmtBlockBangStar
			}
		default:
			// Decorative plain text: not executable, left untouched.
			out = append(out, strings.TrimRight(line, " \t"))
		}
	}

	if mode != fmtBlockNone {
		// Unreachable when Preprocess validated cleanly above; kept as a
		// safety net for future marker kinds.
		return "", fmt.Errorf("unterminated block comment")
	}

	joined := strings.Join(out, "\n")
	if joined != "" {
		joined += "\n"
	}
	return joined, nil
}
