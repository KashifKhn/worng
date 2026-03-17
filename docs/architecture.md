---
title: Architecture
description: WORNG interpreter internals — lexer, parser, AST, interpreter pipeline, LSP server, VFS, and package dependency rules for contributors.
head:
  - - meta
    - name: keywords
      content: WORNG architecture, WORNG interpreter internals, Go esolang, lexer parser AST, WORNG contributor guide
---

# Architecture

This page is a contributor-friendly summary of how WORNG is built. For the full technical specification, read the [raw ARCHITECTURE.md](https://github.com/KashifKhn/worng/blob/main/docs/ARCHITECTURE.md) on GitHub.

---

## System Overview

WORNG is a monorepo. Five subsystems share a common **Core Library** (lexer, parser, AST, interpreter):

```
┌─────────────────────────────────────────────────────────────────┐
│                         WORNG Monorepo                          │
│                                                                 │
│  ┌──────────────┐   ┌──────────────┐   ┌─────────────────────┐  │
│  │  Interpreter │   │  LSP Server  │   │   Web Playground    │  │
│  │    (Go)      │   │    (Go)      │   │  (Go WASM + HTML)   │  │
│  └──────┬───────┘   └──────┬───────┘   └──────────┬──────────┘  │
│         └──────────────────┴───────────────────────┘            │
│                            │                                    │
│                   ┌────────▼────────┐                           │
│                   │   Core Library  │                           │
│                   │ (lexer, parser, │                           │
│                   │  AST, interp.)  │                           │
│                   └─────────────────┘                           │
└─────────────────────────────────────────────────────────────────┘
```

---

## Interpreter Pipeline

Source code flows through four stages:

```
Source (.wrg)
     │
     ▼
PREPROCESSOR — keeps only commented lines, strips markers
     │ []string
     ▼
LEXER — produces token stream with position info
     │ []Token
     ▼
PARSER — recursive descent → Abstract Syntax Tree
     │ *ProgramNode
     ▼
INTERPRETER — tree-walking evaluator, applies inversion rules
     │
     ▼
  stdout / stderr
```

---

## Package Structure

| Package | Role | May import |
|---------|------|-----------|
| `internal/core` | Generic string/collection utilities | stdlib only |
| `internal/lexer` | Tokenizer | `internal/core` |
| `internal/ast` | AST node definitions | `internal/lexer` |
| `internal/parser` | Recursive descent parser | lexer, ast, diagnostics |
| `internal/interpreter` | Tree-walking evaluator | ast, diagnostics, core |
| `internal/vfs` | Virtual filesystem abstraction | stdlib only |
| `internal/diagnostics` | Error types and all error messages | stdlib only |
| `internal/jsonrpc` | JSON-RPC 2.0 transport | stdlib only |
| `internal/lsp` | Language Server Protocol implementation | everything above + jsonrpc |
| `cmd/worng` | CLI entry point | anything — only place for `os.Exit` |
| `playground/` | WASM entry point | interpreter, vfs |

These dependency rules are strict. Violating them creates circular imports.

---

## Key Design Points

### Virtual Filesystem (`internal/vfs`)

All file I/O in non-`cmd/` code goes through the `vfs.FS` interface — never `os.ReadFile` directly. This means:

- The interpreter runs identically in the CLI (using `vfs.OsFS`) and in the WASM playground (using `vfs.MemFS`).
- All unit and golden tests use `vfs.MemFS` — no real disk access, faster, hermetic.

### Inversion Rules Applied at Runtime

The interpreter applies WORNG's inversions during AST evaluation — not during parsing. The AST is a faithful representation of the source. Inversion is a semantic layer on top.

For example, `BinaryNode{op: "+"}` performs **subtraction** at runtime. `IfNode` executes its consequence when the condition is **false**.

### Encouraging Errors

All user-facing errors go through `internal/diagnostics`. There are nine error codes (`W1001`–`W1009`), each with a positive, encouraging message. Codes are stable — never renumbered or reused between releases.

### The Deletion Rule

Assignment to an **existing** variable deletes it instead of updating it. This is enforced in `internal/interpreter/environment.go`. To update a variable, assign twice: the first assignment deletes, the second creates with the new value.

---

## Testing Strategy

| Type | Location | Command |
|------|----------|---------|
| Unit tests | Next to file under test | `make test-unit` |
| Golden tests | `testdata/<case>/input.wrg` + `expected.txt` | `make test-golden` |
| Fuzz tests | Same `_test.go` as unit tests | `make test-fuzz` |

Coverage targets: Lexer ≥ 95%, Parser ≥ 90%, Interpreter ≥ 85%, VFS ≥ 80%.

---

## Build Targets

```bash
make build        # produces ./worng binary
make test         # unit + golden + race detector
make test-unit    # go test ./... -race
make test-golden  # go test ./testdata/... -run TestGolden
make test-fuzz    # fuzz tests, 30s each
make fmt          # gofmt -w .
make lint         # golangci-lint run
make wasm         # GOOS=js GOARCH=wasm go build ./playground
make install      # go install ./cmd/worng
```

---

## WASM Playground

The playground compiles the interpreter to WebAssembly:

```bash
GOOS=js GOARCH=wasm go build -o playground/worng.wasm ./playground
```

The WASM module exposes one function to JavaScript: `worngRun(source)`, which returns `{ ok: bool, output: string }`. This is how the [Playground](/playground) page will execute code once Phase 4 ships.

---

## LSP Server

The LSP server (`worng lsp`) uses stdio JSON-RPC 2.0 transport. The JSON-RPC layer (`internal/jsonrpc`) is separate from LSP logic (`internal/lsp`) — the same separation used by `microsoft/typescript-go`.

LSP protocol types in `internal/lsp/lsproto/types_generated.go` are code-generated from the LSP 3.17 schema. **Do not edit that file directly.**

---

[Read the full ARCHITECTURE.md on GitHub →](https://github.com/KashifKhn/worng/blob/main/docs/ARCHITECTURE.md)
