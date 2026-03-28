package main

import (
	"fmt"
	"os"

	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/interpreter"
	"github.com/KashifKhn/worng/internal/lexer"
	"github.com/KashifKhn/worng/internal/parser"
	"github.com/KashifKhn/worng/internal/vfs"
)

func checkCommand(args []string) int {
	order, jsonOutput, maxErrors, rest, err := parseExecutionFlags(args)
	if err != nil {
		printDiagnostics(os.Stderr, err, vfs.OsFS{}, "", jsonOutput)
		return 2
	}
	args = rest

	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: worng check [--order=btt|ttb] [--json] [--max-errors=N] <file>")
		return 2
	}
	if err := checkFile(vfs.OsFS{}, args[0], order, maxErrors); err != nil {
		printDiagnostics(os.Stderr, err, vfs.OsFS{}, args[0], jsonOutput)
		return 1
	}
	_, _ = fmt.Fprintln(os.Stdout, "OK")
	return 0
}

func checkFile(fs vfs.FS, path string, _ interpreter.ExecutionOrder, maxErrors int) error {
	data, err := fs.ReadFile(path)
	if err != nil {
		return diagnostics.NewFileNotFound(path, err)
	}
	lines := lexer.Preprocess(string(data))
	tokens := lexer.New(joinExecutableLines(lines)).Tokenize()
	p := parser.NewWithFile(tokens, path)
	_, errs := p.Parse()
	if len(errs) > 0 {
		errs = limitErrors(errs, maxErrors)
		return diagnostics.NewErrorList(errs)
	}
	return nil
}
