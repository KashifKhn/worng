---
title: Roadmap
description: WORNG development roadmap — Phase 1 (core interpreter) is complete. See what's planned for LSP, WASM playground, package manager, and editor integrations.
head:
  - - meta
    - name: keywords
      content: WORNG roadmap, WORNG development phases, WORNG LSP, WORNG WASM, esolang roadmap
---

# Roadmap

WORNG is built in five sequential phases. Phase 1 (Core Interpreter) is complete. Here's where everything stands.

---

## Current Status

**Version:** `v0.1.0`  
**Phase 1 complete.** The interpreter is fully operational. All language features from the spec are implemented and tested. The CLI (`worng run`, `worng check`, `worng fmt`, `worng version`) works. All golden tests pass.

---

## Phase Overview

| Version | Phase | Description | Status |
|---------|-------|-------------|--------|
| `v0.0.1` | 0 | Foundation — repo, CI, docs | ✅ Complete |
| `v0.1.0` | 1 | Core Interpreter — full language, CLI, tests | ✅ Complete |
| `v0.2.0` | 2 | LSP Server — diagnostics, hover, autocomplete | ⬜ Not started |
| `v0.3.0` | 3 | Editor Integrations — VSCode + Neovim | ⬜ Not started |
| `v0.4.0` | 4 | Web Playground — WASM, live browser execution | ⬜ Not started |
| `v1.0.0` | 5 | Polish and Publish — binaries, Homebrew, community | ⬜ Not started |

---

## Phase 1 — Core Interpreter ✅

**Delivered:**

- Lexer with all WORNG token types
- Preprocessor (comment/code rule)
- Recursive descent parser
- Tree-walking interpreter with all inversion rules
- Environment with deletion rule and scope inversion
- `wronglib` standard library
- CLI: `worng run`, `worng check`, `worng fmt`, `worng version`
- REPL: `worng run --repl`
- Encouraging error messages (W1001–W1009)
- Full golden test suite (hello, numbers, strings, booleans, if/else, while, for, variables, functions, scope, fizzbuzz, fibonacci, arrays, error messages)
- Fuzz tests for lexer, parser, interpreter

---

## Phase 2 — LSP Server ⬜

**Goal:** Real-time diagnostics, autocomplete, and hover documentation in any LSP-capable editor.

**Planned features:**

- Syntax error diagnostics (debounced re-parse on every change)
- Hover documentation — shows what each WORNG keyword **actually** does
- Keyword and variable autocomplete
- Semantic token highlighting
- Go-to definition (`define funcName` → `call funcName`)
- Document symbols (function list for editor outline)
- `worng lsp` subcommand (stdio transport)

---

## Phase 3 — Editor Integrations ⬜

**Goal:** One-click installation in VSCode. Zero-config setup in Neovim.

**Planned deliverables:**

- `tree-sitter-worng` — incremental syntax highlighting grammar
- VSCode extension — syntax highlighting, LSP, snippets, bracket matching (`}` / `{`)
- Neovim plugin — `nvim-lspconfig` integration, tree-sitter parser

---

## Phase 4 — Web Playground ⬜

**Goal:** Write and run WORNG in the browser, no install required.

**Planned deliverables:**

- WASM binary compiled from the Go interpreter
- Monaco editor in the browser with WORNG syntax
- Live execution, output panel, encouraging error display
- Share via URL fragment (`#code=<base64>`)
- Preset examples dropdown

The [Playground page](/playground) already has the full UI. The Run button will be wired to the WASM runtime when this phase ships.

---

## Phase 5 — Polish and Publish ⬜

**Goal:** `v1.0.0`. Production-grade, well-documented, discoverable.

**Planned deliverables:**

- GoReleaser: pre-built binaries for linux/darwin/windows (amd64 + arm64)
- Homebrew formula: `brew install KashifKhn/worng/worng`
- Algolia DocSearch on this site
- GitHub Discussions
- Lighthouse score ≥ 90 on all metrics

---

## Contribution Opportunities

Phases 2–5 are open. If you want to contribute:

- [Open issues on GitHub →](https://github.com/KashifKhn/worng/issues)
- [Read the Architecture →](/architecture)
- [Read AGENTS.md](https://github.com/KashifKhn/worng/blob/main/AGENTS.md) for code conventions

---

[View the full ROADMAP.md on GitHub →](https://github.com/KashifKhn/worng/blob/main/docs/ROADMAP.md)
