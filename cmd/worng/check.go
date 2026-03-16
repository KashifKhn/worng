package main

import (
	"fmt"
	"os"

	"github.com/KashifKhn/worng/internal/interpreter"
	"github.com/KashifKhn/worng/internal/lexer"
	"github.com/KashifKhn/worng/internal/parser"
	"github.com/KashifKhn/worng/internal/vfs"
)

func checkCommand(args []string) int {
	order := interpreter.OrderBottomToTop
	if len(args) > 0 {
		parsedOrder, consumed, err := parseOrderFlag(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		if consumed {
			order = parsedOrder
			args = args[1:]
		}
	}

	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: worng check [--order=btt|ttb] <file>")
		return 2
	}
	if err := checkFile(vfs.OsFS{}, args[0], order); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	_, _ = fmt.Fprintln(os.Stdout, "OK")
	return 0
}

func checkFile(fs vfs.FS, path string, _ interpreter.ExecutionOrder) error {
	data, err := fs.ReadFile(path)
	if err != nil {
		return err
	}
	lines := lexer.Preprocess(string(data))
	tokens := lexer.New(joinExecutableLines(lines)).Tokenize()
	p := parser.New(tokens)
	_, errs := p.Parse()
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
