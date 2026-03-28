package main

import (
	"fmt"
	"os"
)

var version = "0.1.0"

func main() {
	code := runCLI(os.Args[1:])
	os.Exit(code)
}

func runCLI(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 2
	}

	switch args[0] {
	case "run":
		return runCommand(args[1:])
	case "check":
		return checkCommand(args[1:])
	case "fmt":
		return fmtCommand(args[1:])
	case "lsp":
		return lspCommand()
	case "version":
		fmt.Println(version)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		printUsage()
		return 2
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  worng run [--order=btt|ttb] [--json] [--max-errors=N] <file>")
	fmt.Fprintln(os.Stderr, "  worng run [--order=btt|ttb] --repl")
	fmt.Fprintln(os.Stderr, "  worng check [--order=btt|ttb] [--json] [--max-errors=N] <file>")
	fmt.Fprintln(os.Stderr, "  worng fmt <file>")
	fmt.Fprintln(os.Stderr, "  worng lsp")
	fmt.Fprintln(os.Stderr, "  worng version")
}
