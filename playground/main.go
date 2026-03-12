//go:build js && wasm

// Package main is the WASM entry point for the WORNG web playground.
// Build with: GOOS=js GOARCH=wasm go build -o playground/worng.wasm ./playground
package main

import "syscall/js"

func main() {
	js.Global().Set("worngRun", js.FuncOf(runWorng))
	js.Global().Set("worngCheck", js.FuncOf(checkWorng))
	// Keep the WASM module alive
	select {}
}

func runWorng(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return map[string]any{"ok": false, "output": "No source provided"}
	}
	// TODO(Phase 4): wire up interpreter
	_ = args[0].String()
	return map[string]any{"ok": false, "output": "Not yet implemented (Phase 4)"}
}

func checkWorng(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return map[string]any{"ok": false, "output": "No source provided"}
	}
	// TODO(Phase 4): wire up parser
	_ = args[0].String()
	return map[string]any{"ok": false, "output": "Not yet implemented (Phase 4)"}
}
