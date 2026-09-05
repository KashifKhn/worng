//go:build js && wasm

// Package main is the WASM entry point for the WORNG web playground.
// Build with: GOOS=js GOARCH=wasm go build -o docs/public/worng.wasm ./playground
package main

import (
	"syscall/js"

	"github.com/KashifKhn/worng/internal/interpreter"
	"github.com/KashifKhn/worng/internal/parser"
	"github.com/KashifKhn/worng/internal/play"
)

func main() {
	// The browser's JS stack is far smaller than a native Go stack; deep
	// parser/interpreter recursion inside a js.FuncOf callback overflows it
	// (RangeError) long before our native guards fire. Bound both to stay
	// comfortably inside it while still supporting deep recursion.
	parser.SetMaxExprDepth(300)
	interpreter.SetMaxCallDepth(1000)
	interpreter.SetMaxEvalDepth(12000)

	js.Global().Set("worngRun", js.FuncOf(runWorng))
	js.Global().Set("worngCheck", js.FuncOf(checkWorng))
	// Keep the WASM module alive
	select {}
}

// runWorng implements window.worngRun(source, order?, stdin?).
// Returns { ok, output, diagnostics } where diagnostics is an array of
// { code, key, message, detail, hint, line, column }.
func runWorng(_ js.Value, args []js.Value) any {
	source, order, stdin := readArgs(args)
	res := play.RunSource(source, order, stdin)
	return resultToJS(res)
}

// checkWorng implements window.worngCheck(source).
// Returns { ok, diagnostics } — parse-only, no evaluation.
func checkWorng(_ js.Value, args []js.Value) any {
	if len(args) == 0 || args[0].Type() != js.TypeString {
		return map[string]any{"ok": false, "output": "No source provided"}
	}
	diags := play.CheckSource(args[0].String())
	return map[string]any{
		"ok":          len(diags) == 0,
		"diagnostics": diagsToJS(diags),
	}
}

func readArgs(args []js.Value) (source, order, stdin string) {
	if len(args) > 0 && args[0].Type() == js.TypeString {
		source = args[0].String()
	}
	if len(args) > 1 && args[1].Type() == js.TypeString {
		order = args[1].String()
	}
	if len(args) > 2 && args[2].Type() == js.TypeString {
		stdin = args[2].String()
	}
	return source, order, stdin
}

func resultToJS(res play.Result) map[string]any {
	out := map[string]any{
		"ok":     res.OK,
		"output": res.Output,
	}
	if len(res.Diagnostics) > 0 {
		out["diagnostics"] = diagsToJS(res.Diagnostics)
	} else {
		out["diagnostics"] = []any{}
	}
	return out
}

func diagsToJS(diags []play.Diagnostic) []any {
	out := make([]any, 0, len(diags))
	for _, d := range diags {
		out = append(out, map[string]any{
			"code":    d.Code,
			"key":     d.Key,
			"message": d.Message,
			"detail":  d.Detail,
			"hint":    d.Hint,
			"line":    d.Line,
			"column":  d.Column,
		})
	}
	return out
}
