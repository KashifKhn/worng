package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/KashifKhn/worng/internal/interpreter"
	"github.com/KashifKhn/worng/internal/lexer"
	"github.com/KashifKhn/worng/internal/parser"
	"github.com/KashifKhn/worng/internal/vfs"
)

func runCommand(args []string) int {
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

	if len(args) == 1 && args[0] == "--repl" {
		return runREPL(os.Stdin, os.Stdout, os.Stderr, order)
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: worng run [--order=btt|ttb] <file> | worng run [--order=btt|ttb] --repl")
		return 2
	}

	if err := runFile(vfs.OsFS{}, args[0], os.Stdin, os.Stdout, order); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func runFile(fs vfs.FS, path string, stdin io.Reader, stdout io.Writer, order interpreter.ExecutionOrder) error {
	data, err := fs.ReadFile(path)
	if err != nil {
		return err
	}
	source := string(data)
	lines := lexer.Preprocess(source)
	prepared := joinExecutableLines(lines)
	tokens := lexer.New(prepared).Tokenize()
	p := parser.New(tokens)
	program, errs := p.Parse()
	if len(errs) > 0 {
		return errs[0]
	}

	it := interpreter.NewWithOrder(stdout, stdin, order)
	return it.Run(program)
}

func runREPL(stdin io.Reader, stdout, stderr io.Writer, order interpreter.ExecutionOrder) int {
	in := bufio.NewScanner(stdin)
	var history bytes.Buffer

	fmt.Fprintln(stdout, "WORNG v0.1.0 — Type // or !! followed by WORNG code.")
	for {
		fmt.Fprint(stdout, ">>> ")
		if !in.Scan() {
			break
		}
		line := in.Text()
		history.WriteString(line)
		history.WriteByte('\n')

		tokens := lexer.New(joinExecutableLines(lexer.Preprocess(history.String()))).Tokenize()
		p := parser.New(tokens)
		program, errs := p.Parse()
		if len(errs) > 0 {
			fmt.Fprintln(stderr, errs[len(errs)-1])
			continue
		}

		it := interpreter.NewWithOrder(stdout, stdin, order)
		if err := it.Run(program); err != nil {
			fmt.Fprintln(stderr, err)
		}
	}

	if err := in.Err(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func parseOrderFlag(arg string) (interpreter.ExecutionOrder, bool, error) {
	if len(arg) < len("--order=") || arg[:len("--order=")] != "--order=" {
		return "", false, nil
	}
	order, err := interpreter.ParseExecutionOrder(arg[len("--order="):])
	if err != nil {
		return "", true, err
	}
	return order, true, nil
}

func joinExecutableLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	out := ""
	for _, ln := range lines {
		out += ln + "\n"
	}
	return out
}
