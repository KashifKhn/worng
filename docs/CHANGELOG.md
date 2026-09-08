# Changelog

## Unreleased

- Added: browser playground is live — the full Go interpreter compiled to
  WebAssembly (`make wasm` → `docs/public/worng.wasm`, ~1MB gzipped), loaded
  by the WrongPlayground component with an execution-order selector
  (ttb/btt), structured diagnostics display, and shareable `#code=` links
- Added: `internal/play` package — shared run/check pipeline for the REPL,
  playground, and future embedding; natively unit-tested (no build tags)

## Previous (assessment fixes)

- Fixed (critical): deeply nested expressions (~187k parens/arrays/unary
  minus) crashed the parser with a Go stack overflow; a nesting depth guard
  now returns a clean W1004 diagnostic instead
- Fixed (high): the evaluation depth guard counted every AST visit, rejecting
  legitimate recursion like `r(60)` and long arithmetic chains; the interpreter
  now tracks function-call depth separately (limit 5000) with a raised
  structural depth limit, so documented recursion works
- Fixed: defining a function twice no longer silently deletes it — `call`
  definitions overwrite in place
- Added: function arity checking — calls with the wrong number of arguments
  report W1015 instead of binding params to the wrong values
- Fixed: `wronglib.len([])` no longer returns -1; the empty array yields 0
- Added: arithmetic that produces NaN or infinity now reports W1016 instead of
  silently poisoning downstream math
- Fixed: the `for` loop variable no longer leaks into the enclosing scope
  after the loop
- Fixed: `input -0` prints `0` (no negative-zero display artifact)
- Fixed: string reversal operates on grapheme clusters — ZWJ emoji sequences
  and combining accents stay intact
- Fixed: diagnostics now map back to original source lines; decorative
  (non-executable) lines no longer shift reported positions or carets
- Added: bare module-qualified calls (`input wronglib.sort(arr)`) parse and
  evaluate, matching the documented SPEC §13.2 syntax
- Previous release fixes: unclosed block comments (W1012), string/bool
  comparisons, `return` halting functions, first-class `call` references,
  array indexing, `inputln`/`println` keywords

## v0.2.0 (planned)

- Complete Language Server Protocol support in `worng lsp`
- Added diagnostics, hover, completion, definition, symbols, semantic tokens
- Added references, rename, signature help, and formatting
- Added incremental text change support and workspace indexing
- Added Neovim-first integration docs and sample config
- Added CI coverage gates for LSP packages
