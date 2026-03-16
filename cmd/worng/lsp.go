package main

import (
	"fmt"
	"os"
)

func lspCommand() int {
	// TODO(Phase 2): start LSP server on stdio
	_, _ = fmt.Fprintln(os.Stderr, "worng lsp: not yet implemented (Phase 2)")
	return 1
}
