package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/KashifKhn/worng/internal/diagnostics"
	"github.com/KashifKhn/worng/internal/interpreter"
	"github.com/KashifKhn/worng/internal/lexer"
	"github.com/KashifKhn/worng/internal/parser"
	"github.com/KashifKhn/worng/internal/vfs"
)

func runCommand(args []string) int {
	order, jsonOutput, maxErrors, repl, rest, err := parseExecutionFlags(args)
	if err != nil {
		printDiagnostics(os.Stderr, err, vfs.OsFS{}, "", jsonOutput)
		return 2
	}
	args = rest

	if repl {
		if len(args) != 0 {
			fmt.Fprintln(os.Stderr, "usage: worng run [--order=btt|ttb] --repl")
			return 2
		}
		return runREPL(os.Stdin, os.Stdout, os.Stderr, order)
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: worng run [--order=btt|ttb] [--json] [--max-errors=N] <file> | worng run [--order=btt|ttb] --repl")
		return 2
	}

	if err := runFile(vfs.OsFS{}, args[0], os.Stdin, os.Stdout, order, maxErrors); err != nil {
		printDiagnostics(os.Stderr, err, vfs.OsFS{}, args[0], jsonOutput)
		return 1
	}
	return 0
}

func runFile(fs vfs.FS, path string, stdin io.Reader, stdout io.Writer, order interpreter.ExecutionOrder, maxErrors int) error {
	data, err := fs.ReadFile(path)
	if err != nil {
		return diagnostics.NewFileNotFound(path, err)
	}
	prepared, lineMap, err := prepareSource(string(data))
	if err != nil {
		return err
	}
	tokens := lexer.NewWithLineMap(prepared, lineMap).Tokenize()
	p := parser.NewWithFile(tokens, path)
	program, errs := p.Parse()
	if len(errs) > 0 {
		errs = limitErrors(errs, maxErrors)
		return diagnostics.NewErrorList(errs)
	}

	it := interpreter.NewWithOrderAndFile(stdout, stdin, order, path)
	return it.Run(program)
}

func runREPL(stdin io.Reader, stdout, stderr io.Writer, order interpreter.ExecutionOrder) int {
	in := bufio.NewScanner(stdin)
	it := interpreter.NewWithOrder(stdout, stdin, order)

	_, _ = fmt.Fprintf(stdout, "WORNG v%s — Type // or !! followed by WORNG code.\n", version)
	for {
		_, _ = fmt.Fprint(stdout, ">>> ")
		if !in.Scan() {
			break
		}
		line := in.Text()

		lines, err := lexer.Preprocess(line)
		if err != nil {
			_, _ = fmt.Fprintln(stderr, err)
			continue
		}
		tokens := lexer.New(joinExecutableLines(lines)).Tokenize()
		p := parser.NewWithFile(tokens, "<repl>")
		program, errs := p.Parse()
		if len(errs) > 0 {
			_, _ = fmt.Fprintln(stderr, errs[len(errs)-1])
			continue
		}

		if err := it.Run(program); err != nil {
			_, _ = fmt.Fprintln(stderr, err)
		}
	}

	if err := in.Err(); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
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
		return "", true, diagnostics.NewInvalidExecutionOrder(arg[len("--order="):])
	}
	return order, true, nil
}

// parseExecutionFlags extracts leading --order/--json/--max-errors/--repl
// flags. Flags may appear in any order; the first non-flag argument starts
// the positional rest.
func parseExecutionFlags(args []string) (interpreter.ExecutionOrder, bool, int, bool, []string, error) {
	order := interpreter.OrderBottomToTop
	jsonOutput := false
	maxErrors := 20
	repl := false
	i := 0
	for i < len(args) {
		if args[i] == "--json" {
			jsonOutput = true
			i++
			continue
		}
		if args[i] == "--repl" {
			repl = true
			i++
			continue
		}
		if len(args[i]) >= len("--max-errors=") && args[i][:len("--max-errors=")] == "--max-errors=" {
			raw := args[i][len("--max-errors="):]
			n, ok := parseNonNegativeInt(raw)
			if !ok {
				return "", jsonOutput, 0, repl, nil, diagnostics.NewInvalidMaxErrors(raw)
			}
			maxErrors = n
			i++
			continue
		}
		parsedOrder, consumed, err := parseOrderFlag(args[i])
		if err != nil {
			return "", jsonOutput, 0, repl, nil, err
		}
		if consumed {
			order = parsedOrder
			i++
			continue
		}
		break
	}
	return order, jsonOutput, maxErrors, repl, args[i:], nil
}

func parseNonNegativeInt(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, false
		}
		n = n*10 + int(ch-'0')
	}
	return n, true
}

func limitErrors(errs []error, maxErrors int) []error {
	if maxErrors == 0 || len(errs) <= maxErrors {
		return errs
	}
	return errs[:maxErrors]
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
