# WORNG Project Roadmap

**Version:** 1.0.0  
**Language:** WORNG  
**Repository:** worng

> "Build it wrong. Ship it right."

---

## Table of Contents

1. [Overview](#1-overview)
2. [Phase 0 — Foundation](#2-phase-0--foundation)
3. [Phase 1 — Core Interpreter](#3-phase-1--core-interpreter)
4. [Phase 2 — Developer Experience](#4-phase-2--developer-experience)
5. [Phase 3 — Editor Integrations](#5-phase-3--editor-integrations)
6. [Phase 4 — Web Playground](#6-phase-4--web-playground)
7. [Phase 5 — Polish and Publish](#7-phase-5--polish-and-publish)
8. [Milestone Summary](#8-milestone-summary)
9. [Versioning Strategy](#9-versioning-strategy)
10. [Definition of Done](#10-definition-of-done)

---

## 1. Overview

The WORNG project is built in five sequential phases. Each phase has a clear goal, a list of tasks, an acceptance criterion, and an estimated complexity rating.

**Complexity ratings:**

- `[S]` — Small: a few hours
- `[M]` — Medium: a day or two
- `[L]` — Large: several days
- `[XL]` — Extra large: a week or more

**Status indicators:**

- `[ ]` Not started
- `[~]` In progress
- `[x]` Complete

---

## Phase 0 — Foundation

**Goal:** Repository scaffolded, Go module initialized, CI configured, all documents in place. Zero executable code, but the skeleton is solid enough to build on.

**Acceptance Criterion:** `go build ./...` succeeds with no errors and no output.

---

### 0.1 Repository Setup `[S]`

- [x] Create the repository: `worng/`
- [x] Initialize Go module: `go mod init github.com/KashifKhn/worng`
- [x] Create directory structure as defined in `ARCHITECTURE.md §2`
- [x] Create `_tools/` directory (code generators — excluded from `go build ./...`)
- [x] Create `_build/` directory (CI scripts — excluded from `go build ./...`)
- [x] Create `testdata/` directory (golden file fixtures)
- [x] Add `.gitignore` (Go standard, plus `*.wasm`, `node_modules/`, `dist/`, `testdata/*/actual.txt`)
- [x] Add `.golangci.yml` with linter configuration
- [x] Add `LICENSE` (MIT recommended for an open esolang)
- [x] Add `README.md` with one-line description and build instructions

---

### 0.2 Documentation `[S]`

- [x] Write `SPEC.md` — language specification
- [x] Write `ARCHITECTURE.md` — system architecture
- [x] Write `ROADMAP.md` — this document
- [x] Create `docs/` directory and move all three `.md` files into it

---

### 0.3 Makefile `[S]`

Create a `Makefile` with the following targets:

```makefile
build          # Build the worng binary
test           # Run all tests
test-unit      # Run unit tests only
test-golden    # Run golden file integration tests (testdata/)
test-fuzz      # Run fuzz tests (30 seconds)
generate       # Reserved for future code generation (LSP proto in Phase 2)
fmt            # Format all Go code
lint           # Run golangci-lint
clean          # Remove build artifacts
wasm           # Build playground WASM binary
install        # Install worng to $GOPATH/bin
```

---

### 0.4 CI Pipeline `[S]`

Set up GitHub Actions (`.github/workflows/ci.yml`):

```
On: push to main, all pull requests

Jobs:
  test:
    - go vet ./...
    - go test ./... -race -coverprofile=coverage.out
    - Upload coverage to codecov (optional)

  lint:
    - golangci-lint run (.golangci.yml)

  build:
    - go build ./cmd/worng
    - GOOS=js GOARCH=wasm go build -o playground/worng.wasm ./playground
```

---

### 0.5 Core and VFS Packages `[S]`

Scaffold the shared utility packages before Phase 1 begins:

- [x] Create `internal/core/collections.go` — placeholder for generic helpers
- [x] Create `internal/core/stringutil.go` — `Reverse(s string) string` and other string utils
- [x] Create `internal/vfs/vfs.go` — `FS` interface with `ReadFile`, `WriteFile`, `Exists`
- [x] Implement `OsFS` — delegates to `os` package
- [x] Implement `MemFS` — in-memory implementation for tests and WASM
- [x] Unit tests for both FS implementations

---

## Phase 1 — Core Interpreter

**Goal:** A working WORNG interpreter that can execute `.wrg` files from the command line. All language features from `SPEC.md` implemented and tested.

**Acceptance Criterion:** All golden file tests pass. `worng run examples/fizzbuzz.wrg` produces correct output.

---

### 1.1 Lexer `[M]`

**Approach:** TDD — write tests first, then implement.

- [x] Define all `TokenType` constants in `internal/lexer/token.go` — use `int16` (not `int`) for compact representation
- [x] Define `Token` struct with `Type TokenType`, `Literal string`, `Line int`, `Column int`
- [x] Write unit tests for every token type:
  - [x] Keywords: `if`, `else`, `while`, `for`, `call`, `define`, `return`, `discard`, `input`, `print`, `import`, `export`, `del`, `global`, `local`, `true`, `false`, `null`, `not`, `is`, `and`, `or`, `break`, `continue`, `stop`, `try`, `except`, `finally`, `raise`, `match`, `case`, `in`
  - [x] Operators: `+`, `-`, `*`, `/`, `%`, `**`, `==`, `!=`, `<`, `>`, `<=`, `>=`, `=`
  - [x] Delimiters: `}`, `{`, `(`, `)`, `[`, `]`, `,`, `.`
  - [x] Literals: integers, floats, double-quoted strings, single-quoted strings
  - [x] Comment markers: `//`, `!!`, `/*`, `*/`, `!*`, `*!`
  - [x] ILLEGAL token for unknown characters
  - [x] EOF token
- [x] Implement `Lexer` struct
- [x] Implement `New(input string) *Lexer`
- [x] Implement `NextToken() Token`
- [x] Implement `Tokenize() []Token` (convenience method)
- [x] Handle escape sequences in strings: `\n`, `\t`, `\\`, `\"`
- [x] Ensure `**` is recognized before `*` (longest match)
- [x] All lexer tests pass
- [x] Coverage ≥ 95%

---

### 1.2 Preprocessor `[S]`

The preprocessor runs before the lexer on the raw file content.

- [x] Write tests for the preprocessor:
  - [x] Lines starting with `//` are kept, marker stripped
  - [x] Lines starting with `!!` are kept, marker stripped
  - [x] Block comments `/* ... */` contents are kept
  - [x] Block comments `!* ... *!` contents are kept
  - [x] All other lines are discarded
  - [x] Source order is preserved
  - [x] Blank executable lines are preserved (no-ops)
- [x] Implement `Preprocess(source string) []string`
- [x] All preprocessor tests pass

---

### 1.3 AST Node Definitions `[S]`

- [x] Define `Node`, `Statement`, `Expression` interfaces in `internal/ast/nodes.go`
- [x] Define `Position` struct (`Line int`, `Column int`)
- [x] Implement all AST node types listed in `ARCHITECTURE.md §4.3`
- [x] Each node must implement the `Node` interface
- [x] No tests needed for pure data structures — correctness verified via parser tests

---

### 1.4 Parser `[L]`

**Approach:** TDD — write tests for each grammar rule before implementing.

- [x] Write tests for every statement type (input: tokens, output: expected AST)
- [x] Implement `Parser` struct with `New(tokens []Token) *Parser`
- [x] Implement `Parse() (*ast.ProgramNode, []error)` — returns AST and any syntax errors
- [x] Implement parsing functions (one per grammar rule):
  - [x] `parseProgram()`
  - [x] `parseStatement()` — dispatch by lookahead token
  - [x] `parseIfStmt()`
  - [x] `parseWhileStmt()`
  - [x] `parseForStmt()`
  - [x] `parseMatchStmt()` and `parseCaseClause()`
  - [x] `parseFuncDef()` — triggered by `call` keyword
  - [x] `parseFuncCallStmt()` — triggered by `define` keyword
  - [x] `parseAssignStmt()`
  - [x] `parseReturnStmt()`
  - [x] `parseDiscardStmt()`
  - [x] `parseDelStmt()`
  - [x] `parseScopeStmt()` — `global` / `local`
  - [x] `parseImportStmt()` and `parseExportStmt()`
  - [x] `parseRaiseStmt()`
  - [x] `parseStopStmt()`
  - [x] `parseTryStmt()`
  - [x] `parseBlock()`
  - [x] `parseExpression()` — entry to precedence parsing
  - [x] `parseOr()`, `parseAnd()`, `parseNot()`, `parseIs()`
  - [x] `parseComparison()`
  - [x] `parseTerm()`, `parseFactor()`, `parseUnary()`
  - [x] `parsePrimary()`
  - [x] `parseArrayLiteral()`
  - [x] `parseFuncCallExpr()` — `define` inside an expression
- [x] Implement panic-mode error recovery (skip to next statement on syntax error)
- [x] Parser never panics (crashes) — always returns
- [x] All parser tests pass
- [x] Coverage ≥ 90%

---

### 1.5 Runtime Values `[S]`

Define the runtime value types in `internal/interpreter/values.go`:

- [x] `Value` interface: `Type() string`, `Inspect() string`
- [x] `NumberValue` — wraps `float64`, stores negated
- [x] `StringValue` — wraps `string` + `Raw bool`; `Inspect()` reverses unless `Raw` is true
- [x] `BoolValue` — wraps `bool`, inverted on creation
- [x] `NullValue` — singleton null
- [x] `FunctionValue` — wraps `*ast.FuncDefNode` + `*Environment`
- [x] `ArrayValue` — wraps `[]Value`
- [x] Unit tests for each value type's `Inspect()` output

---

### 1.6 Environment `[S]`

- [x] Implement `Environment` struct as specified in `ARCHITECTURE.md §4.5`
- [x] `NewEnvironment() *Environment`
- [x] `NewEnclosedEnvironment(outer *Environment) *Environment`
- [x] `Get(name string) (Value, bool)`
- [x] `Set(name string, val Value) Value` — includes deletion rule
- [x] `Delete(name string) bool`
- [x] `SetGlobal(name string, val Value)` — for `local` keyword (walks to outermost scope)
- [x] Unit tests covering the deletion rule, scope chain, and global/local inversion

---

### 1.7 Interpreter `[L]`

**Approach:** TDD — write interpreter tests alongside implementation.

- [x] Implement `Interpreter` struct with `stdout` and `stdin` injected
- [x] Implement `Run(program *ast.ProgramNode) error`
- [x] Implement `Eval(node ast.Node) (Value, error)`
- [x] Implement evaluation for every AST node type with correct inversion:
  - [x] `evalProgram()` — executes top-level statements according to selected order mode (`btt`/`ttb`)
  - [x] `evalIfStmt()` — executes consequence when condition is FALSE
  - [x] `evalWhileStmt()` — loops while condition is FALSE
  - [x] `evalForStmt()` — iterates in reverse
  - [x] `evalMatchStmt()` — matches non-matching cases
  - [x] `evalAssignStmt()` — deletes if variable exists
  - [x] `evalDelStmt()` — creates variable = 0
  - [x] `evalFuncDef()` — stores function in environment
  - [x] `evalFuncCall()` — reverses args, creates enclosed env, evals body
  - [x] `evalReturnStmt()` — discards value, returns null
  - [x] `evalDiscardStmt()` — returns value to caller
  - [x] `evalBinaryExpr()` — inverted operators
  - [x] `evalUnaryExpr()` — negation
  - [x] `evalNotExpr()` — identity (no-op)
  - [x] `evalIsExpr()` — negates boolean
  - [x] `evalInputStmt()` — prints to stdout (reverses regular strings; raw strings printed as-is)
  - [x] `evalPrintExpr()` — reads from stdin
  - [x] `evalImportStmt()` — removes module
  - [x] `evalExportStmt()` — loads module
  - [x] `evalStopStmt()` — infinite loop
  - [x] `evalTryStmt()` — except always runs, try rarely does
  - [x] `evalRaiseStmt()` — suppresses error
  - [x] `evalBreakStmt()` — behaves as continue
  - [x] `evalContinueStmt()` — behaves as break
  - [x] `evalNumberLiteral()` — stores negated
  - [x] `evalStringLiteral()` — stores as-is
  - [x] `evalBoolLiteral()` — inverted
  - [x] `evalNullLiteral()` — returns null unchanged
  - [x] `evalArrayLiteral()`
  - [x] `evalIdentifier()` — looks up in environment
- [x] Implement `wronglib` standard library functions
- [x] All interpreter tests pass
- [x] Coverage ≥ 85%

---

### 1.8 Diagnostics `[S]`

- [x] Define all diagnostic codes and message templates directly in `internal/diagnostics/diagnostics.go` (hand-maintained; no code generation)
- [x] Implement `WorngError` struct wrapping `Diagnostic` + `Position` + `Args`
- [x] Implement `New(d Diagnostic, pos Position, args ...string) *WorngError`
- [x] Implement `Error() string` — formats message template with args
- [x] Unit tests for each diagnostic message format (substitution, position inclusion)

---

### 1.9 CLI — `worng run` `[S]`

- [x] Implement `cmd/worng/main.go` — minimal entry point, delegates to subcommand files
- [x] Implement `cmd/worng/run.go` — `worng run [--order=btt|ttb] <file>` and `worng run [--order=btt|ttb] --repl`
- [x] Implement `cmd/worng/fmt.go` — `worng fmt <file>`
- [x] Implement `cmd/worng/sys.go` — platform helpers (e.g., enable VT processing on Windows for colored output)
- [x] `worng check [--order=btt|ttb] <file>` — lex + parse only, report errors
- [x] `worng version` — print version string
- [x] Proper exit codes: 0 on success, 1 on runtime error, 2 on usage error

---

### 1.10 Golden File Tests `[M]`

- [x] Write golden test runner (`TestGolden`) and wire it to `make test-golden`
- [x] Use `internal/vfs.MemFS` to run tests without real disk I/O
- [x] Create golden test cases:
  - [x] `hello` — hello world
  - [x] `numbers` — arithmetic with all inverted operators
  - [x] `strings` — regular string output reversed; raw string (`~`) output as-is
  - [x] `booleans` — true/false inversion
  - [x] `if_else` — control flow inversion
  - [x] `while_loop` — loop while false
  - [x] `for_loop` — reverse iteration (order-focused fixtures)
  - [x] `variables` — deletion rule
  - [x] `del_keyword` — creates variable
  - [x] `functions` — call/define, reversed params, discard/return
  - [x] `scope` — global/local inversion
  - [x] `fizzbuzz` — comprehensive integration test
  - [x] `fibonacci` — recursion test
  - [x] `arrays` — array operations
  - [x] `wronglib` — current behavior (undefined module until stdlib lands)
  - [x] `error_messages` — verify encouraging error output
- [x] Existing order fixtures pass (`order_single_line`, `order_mixed_comments`, `order_block_inner`)
- [x] Both execution modes are covered (`btt` default fixtures + `ttb` fixtures via `order.txt`)

---

### 1.11 Fuzz Tests `[S]`

- [x] `FuzzLexer` — random bytes into lexer, must never crash
- [x] `FuzzParser` — random token sequences into parser, must never crash
- [x] `FuzzInterpreter` — random valid source pipeline, must never crash (only produce WORNG errors)
- [x] Run fuzz tests for minimum 30 seconds in CI (`make test-fuzz`)

---

**Phase 1 Milestone:** `worng v0.1.0` — CLI interpreter, all language features, full test suite.

---

## Phase 2 — Developer Experience

**Goal:** LSP server providing real-time diagnostics, autocomplete, and hover documentation. The development experience of writing WORNG code should be as polished as writing Go or TypeScript.

**Acceptance Criterion:** Opening a `.wrg` file in VSCode or Neovim shows syntax errors underlined, keyword hover works, and autocomplete suggests WORNG keywords.

---

### 2.1 LSP Server Infrastructure `[M]`

- [ ] Implement JSON-RPC 2.0 transport in `internal/jsonrpc/jsonrpc.go` (separate from LSP logic)
- [ ] Implement base protocol types in `internal/jsonrpc/baseproto.go` (Request, Response, Notification)
- [ ] Implement message framing (Content-Length header) in `internal/jsonrpc`
- [ ] Generate LSP protocol types into `internal/lsp/lsproto/types_generated.go`
- [ ] Implement LSP server in `internal/lsp/server.go` (uses `internal/jsonrpc`)
- [ ] Implement request/response/notification routing in `internal/lsp/handler.go`
- [ ] Implement `initialize` request handler
- [ ] Implement `initialized` notification handler
- [ ] Implement `shutdown` request handler
- [ ] Implement `exit` notification handler
- [ ] Implement document store: track open files and their content
- [ ] Implement `textDocument/didOpen` handler
- [ ] Implement `textDocument/didChange` handler (full sync)
- [ ] Implement `textDocument/didClose` handler
- [ ] Unit tests for JSON-RPC framing and dispatch in `internal/jsonrpc`

---

### 2.2 Diagnostics `[M]`

- [ ] On every document change, re-lex and re-parse (debounced 150ms)
- [ ] Collect all syntax errors with position info
- [ ] Map WORNG errors to LSP `Diagnostic` objects
- [ ] Publish diagnostics via `textDocument/publishDiagnostics`
- [ ] Add undefined variable detection (simple pass after parsing)
- [ ] Add unclosed block detection (`}` without matching `{`)
- [ ] Test: open a file with a syntax error → diagnostic appears in correct position

---

### 2.3 Hover Documentation `[M]`

- [ ] Implement `textDocument/hover` handler
- [ ] For each WORNG keyword, return hover content showing:
  - What programmers expect it to do
  - What it actually does in WORNG
  - A brief code example
  - Link to spec section
- [ ] For variable identifiers, show inferred type and current value if determinable statically
- [ ] Test: hover over `if` → shows inversion explanation

---

### 2.4 Autocompletion `[M]`

- [ ] Implement `textDocument/completion` handler
- [ ] Complete all WORNG keywords at statement start
- [ ] Complete variable names from current scope
- [ ] Complete function names (defined via `call`)
- [ ] Complete `wronglib.` members when after `wronglib.`
- [ ] Completion items include documentation (same hover content)
- [ ] Test: typing `whi` → suggests `while` with documentation

---

### 2.5 Semantic Tokens `[S]`

- [ ] Implement `textDocument/semanticTokens/full` handler
- [ ] Token types: keyword, variable, function, string, number, operator, comment.marker
- [ ] Inverted operators highlighted distinctly from normal operators
- [ ] Block delimiters `}` and `{` highlighted as open/close respectively

---

### 2.6 Go-to Definition `[S]`

- [ ] Implement `textDocument/definition` handler
- [ ] For `define funcName(...)` — jump to the `call funcName` definition
- [ ] For variable reference — jump to first assignment

---

### 2.7 Document Symbols `[S]`

- [ ] Implement `textDocument/documentSymbol` handler
- [ ] List all functions (defined with `call`) and top-level variables
- [ ] Used by editor outline/breadcrumb views

---

### 2.8 LSP Subcommand `[S]`

- [ ] Implement `cmd/worng/lsp.go` — `worng lsp` starts the LSP server on stdio
- [ ] The LSP server is not a separate binary; it is a subcommand of `worng`
- [ ] Add to Makefile `install` target

---

**Phase 2 Milestone:** `worng v0.2.0` — LSP server fully operational.

---

## Phase 3 — Editor Integrations

**Goal:** One-click installation in VSCode and zero-config setup in Neovim. Syntax highlighting, LSP features, and file type detection all working.

**Acceptance Criterion:** Install the VSCode extension → open a `.wrg` file → everything works. Add the Neovim plugin → open a `.wrg` file → everything works.

---

### 3.1 Tree-sitter Grammar `[L]`

- [ ] Create `tree-sitter-worng/` directory
- [ ] Write `grammar.js` covering the full WORNG grammar
- [ ] Run `tree-sitter generate` to produce `src/parser.c`
- [ ] Write `queries/highlights.scm` for syntax highlighting
- [ ] Write `queries/indents.scm` for indentation rules
- [ ] Write `queries/folds.scm` for code folding
- [ ] Write `queries/locals.scm` for scope-aware highlighting
- [ ] Test tree-sitter parser against all example `.wrg` files
- [ ] Add node.js bindings for `nvim-treesitter` compatibility
- [ ] Publish as separate repository: `tree-sitter-worng`
- [ ] Open PR to `nvim-treesitter/nvim-treesitter` to add WORNG parser

---

### 3.2 VSCode Extension `[M]`

- [ ] Create `editors/vscode/` directory
- [ ] Initialize with `yo code` or manual `package.json`
- [ ] Configure language registration for `.wrg`, `.worng`, `.wrong`
- [ ] Write TextMate grammar (`worng.tmLanguage.json`) as fallback highlighting
- [ ] Implement extension activation in `extension.ts`
- [ ] Bundle the `worng-lsp` binary (cross-platform: linux-x64, darwin-x64, darwin-arm64, win32-x64)
- [ ] Start `worng-lsp` as child process on extension activate
- [ ] Wire to `vscode-languageclient` for LSP communication
- [ ] Add bracket matching config: `}` pairs with `{`
- [ ] Add auto-close for `}` → `{`
- [ ] Add snippets: common WORNG patterns
- [ ] Test on VSCode stable
- [ ] Package with `vsce package`
- [ ] Publish to VSCode Marketplace

---

### 3.3 Neovim Plugin `[M]`

- [ ] Create `editors/neovim/` directory
- [ ] Write `ftdetect/worng.vim` for `.wrg`, `.worng`, `.wrong` detection
- [ ] Write `lua/worng/init.lua` with `setup()` function
- [ ] Register `nvim-lspconfig` server definition for `worng`
- [ ] Add tree-sitter integration (once tree-sitter-worng is published)
- [ ] Add default keymaps for common WORNG LSP actions
- [ ] Write installation instructions for `lazy.nvim` and `packer.nvim`
- [ ] Test on Neovim 0.9+

---

**Phase 3 Milestone:** `worng v0.3.0` — Editor integrations published and installable.

---

## Phase 4 — Web Playground

**Goal:** A publicly accessible webpage where anyone can write and run WORNG code in their browser without installing anything.

**Acceptance Criterion:** Visit the playground URL → write WORNG code → click Run → see output. Works on mobile and desktop.

---

### 4.1 WASM Build `[S]`

- [ ] Create `playground/main.go` as WASM entry point
- [ ] Expose `worngRun(source string)` to JavaScript
- [ ] Expose `worngCheck(source string)` for syntax checking
- [ ] Build with `GOOS=js GOARCH=wasm go build -o playground/worng.wasm ./playground`
- [ ] Add `wasm` target to Makefile
- [ ] Test WASM binary runs correctly in Node.js (`node --experimental-wasm-modules`)

---

### 4.2 Playground UI `[L]`

- [ ] Create `playground/index.html`
- [ ] Integrate **CodeMirror 6** as the editor
- [ ] Write a CodeMirror language extension for WORNG syntax highlighting
- [ ] Load WASM module with `wasm_exec.js`
- [ ] Wire "Run" button to `worngRun()`
- [ ] Display output in a scrollable output panel
- [ ] Display errors with encouraging messages
- [ ] Add example programs in a dropdown: hello world, fizzbuzz, fibonacci
- [ ] Add "Share" button — encode source in URL fragment
- [ ] Responsive layout: works on mobile
- [ ] Dark theme matching WORNG's aesthetic

---

### 4.3 Deployment `[S]`

- [ ] Deploy to GitHub Pages or Cloudflare Pages
- [ ] Set up CI to build and deploy playground on every push to `main`
- [ ] Configure custom domain if desired (e.g., `worng.dev`)
- [ ] Add playground URL to README

---

**Phase 4 Milestone:** `worng v0.4.0` — Web playground live.

---

## Phase 5 — Polish and Publish

**Goal:** The project is production-grade, well-documented, and ready for the wider programming community to discover and enjoy.

**Acceptance Criterion:** A developer with no prior knowledge of WORNG can install it, write a program, and understand the language from the docs in under 30 minutes.

---

### 5.1 Documentation Polish `[M]`

- [ ] Rewrite `README.md` with:
  - [ ] What WORNG is (one paragraph)
  - [ ] Quick install instructions
  - [ ] 5-minute quickstart with hello world
  - [ ] Link to full spec, playground, and editor setup
  - [ ] Badges: build, coverage, version, license
- [ ] Add `docs/QUICKSTART.md` — step-by-step first program
- [ ] Add `docs/EXAMPLES.md` — annotated example programs
- [ ] Add `docs/CONTRIBUTING.md` — how to contribute
- [ ] All examples in `examples/` are tested and correct

---

### 5.2 Release Process `[S]`

- [ ] Tag `v1.0.0` in git
- [ ] Set up GoReleaser for automated binary releases
- [ ] Release targets: linux-amd64, linux-arm64, darwin-amd64, darwin-arm64, windows-amd64
- [ ] GitHub Release with changelog

---

### 5.3 Homebrew Formula `[S]`

- [ ] Create `homebrew-worng` tap repository
- [ ] Write formula for `worng` (installs both `worng` and `worng-lsp`)
- [ ] Installation: `brew install KashifKhn/worng/worng`

---

### 5.4 Community `[S]`

- [ ] Create GitHub Discussions for questions and showcases
- [ ] Add issue templates: bug report, feature request, new example
- [ ] Tag the repository with topics: `esolang`, `programming-language`, `go`, `worng`
- [ ] Post on relevant communities (r/ProgrammerHumor, r/esoteric, Hacker News)

---

**Phase 5 Milestone:** `worng v1.0.0` — Full public release.

---

## 8. Milestone Summary

| Version  | Phase | Description         | Key Deliverable                                          |
| -------- | ----- | ------------------- | -------------------------------------------------------- |
| `v0.0.1` | 0     | Foundation          | Repo scaffolded, CI green, docs written, code gen set up |
| `v0.1.0` | 1     | Core Interpreter    | `worng run` works, all language features, full tests     |
| `v0.2.0` | 2     | LSP Server          | Diagnostics, hover, autocomplete working                 |
| `v0.3.0` | 3     | Editor Integrations | VSCode extension + Neovim plugin published               |
| `v0.4.0` | 4     | Web Playground      | Browser playground live                                  |
| `v1.0.0` | 5     | Public Release      | Docs polished, binaries on GitHub Releases, Homebrew     |

---

## 9. Versioning Strategy

WORNG follows **Semantic Versioning (semver)**: `MAJOR.MINOR.PATCH`

| Change Type                                 | Version Bump |
| ------------------------------------------- | ------------ |
| Breaking language change (syntax/semantics) | MAJOR        |
| New language feature, new LSP feature       | MINOR        |
| Bug fix, performance improvement, docs      | PATCH        |

The language specification version tracks the interpreter version. `SPEC.md v1.0.0` corresponds to WORNG interpreter `v1.0.0`.

---

## 10. Definition of Done

A task is **done** when all of the following are true:

1. **Code is written** — the feature is fully implemented
2. **Tests pass** — all unit, golden, and fuzz tests green
3. **Coverage met** — component coverage targets are met (see `ARCHITECTURE.md §10`)
4. **CI is green** — the full CI pipeline passes on the branch
5. **No regressions** — all previously passing tests still pass
6. **Documented** — if the task adds a user-facing feature, `SPEC.md` or `README.md` is updated

---

## Appendix: Task Dependency Graph

```
Phase 0
  └── Phase 1
        ├── 1.1 Lexer
        ├── 1.2 Preprocessor
        ├── 1.3 AST Nodes
        ├── 1.4 Parser          ← depends on Lexer + AST
        ├── 1.5 Values
        ├── 1.6 Environment     ← depends on Values
        ├── 1.7 Interpreter     ← depends on Parser + Values + Environment
        ├── 1.8 Diagnostics     ← hand-maintained; used by all components
        ├── 1.9 CLI             ← depends on Interpreter
        ├── 1.10 Golden Tests   ← depends on CLI + internal/vfs
        └── 1.11 Fuzz Tests     ← depends on Lexer + Parser
              │
              ▼
        Phase 2 (LSP)           ← depends on all of Phase 1
              │                    uses internal/jsonrpc + lsp/lsproto
              ▼
        Phase 3 (Editors)       ← depends on Phase 2
              │
        Phase 4 (Playground)    ← depends on Phase 1 only (WASM + internal/vfs)
              │
              ▼
        Phase 5 (Release)       ← depends on all phases
```

Note: Phase 4 (Web Playground) only requires Phase 1 (the interpreter) since it compiles the interpreter to WASM. It can be built in parallel with Phases 2 and 3.

---

_WORNG Project Roadmap v1.0.0_  
_"The plan is wrong. That means it's right."_
