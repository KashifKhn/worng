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

func formatFile(fs vfs.FS, path string) error {
	data, err := fs.ReadFile(path)
	if err != nil {
		return err
	}

	// Minimal formatter for Phase 1: normalize executable lines.
	lines := lexer.Preprocess(string(data))
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	out := strings.Join(lines, "\n")
	if out != "" {
		out += "\n"
	}
	return fs.WriteFile(path, []byte(out))
}
