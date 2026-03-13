# AGENTS.md

## Repository

Module: `github.com/KashifKhn/worng`  
Language: Go (interpreter for the WORNG esoteric programming language)  
Docs: `docs/SPEC.md`, `docs/ARCHITECTURE.md`, `docs/ROADMAP.md`, `docs/WEBSITE.md`

---

## Commands

```bash
# Build
make build                        # produces ./worng binary
go build ./...                    # verify all packages compile (use this to check for errors)

# Test
make test                         # unit + golden + race detector
make test-unit                    # go test ./... -race
make test-golden                  # go test ./testdata/... -run TestGolden
make test-fuzz                    # fuzz tests, 30s each target

# Run a single test package
go test ./internal/lexer/... -v
go test ./internal/parser/... -v
go test ./internal/interpreter/... -v

# Run a single named test
go test ./internal/lexer/... -run TestLexer_NumberToken -v
go test ./internal/parser/... -run TestParseIfStmt -v -count=1
go test ./internal/interpreter/... -run TestEvalBinary -v -count=1

# Run a single fuzz target (30s)
go test ./internal/lexer/... -fuzz=FuzzLexer -fuzztime=30s
go test ./internal/parser/... -fuzz=FuzzParser -fuzztime=30s
go test ./internal/interpreter/... -fuzz=FuzzInterpreter -fuzztime=30s

# Code generation
make generate                     # (reserved for Phase 2 LSP proto generation; currently a no-op)

# Format + lint
make fmt                          # gofmt -w .
make lint                         # golangci-lint run

# WASM playground build
make wasm                         # GOOS=js GOARCH=wasm go build -o playground/worng.wasm ./playground

# Install
make install                      # go install ./cmd/worng
```

---

## Project Layout

```
cmd/worng/          main.go run.go fmt.go lsp.go sys.go
internal/
  core/             stringutil.go collections.go     (no deps on other internal pkgs)
  lexer/            token.go lexer.go
  ast/              nodes.go
  parser/           parser.go
  diagnostics/      diagnostics.go
  interpreter/      interpreter.go environment.go values.go builtins.go
  vfs/              vfs.go
  jsonrpc/          jsonrpc.go baseproto.go
  lsp/              server.go handler.go
    lsproto/        types_generated.go               (DO NOT EDIT — generated)
playground/         main.go                          (build tag: js && wasm)
testdata/           golden test fixtures
_build/             CI scripts                       (excluded from go build ./...)
docs/               SPEC.md ARCHITECTURE.md ROADMAP.md WEBSITE.md
```

`_build/` uses Go's `_` prefix convention — it is never compiled by `go build ./...`.

---

## WORNG Language Rules

These rules must be correctly implemented by every component. Read `docs/SPEC.md` for the full specification.

### Execution Model

- Programs execute **bottom to top** — the preprocessor reverses the order of executable lines
- Only **commented lines** execute; all other lines are silently ignored
- Executable line markers: `//`, `!!` (single-line); `/* ... */`, `!* ... *!` (block)
- Block comments do **not** nest

### Inversion Rules (implement these exactly)

| Written | Actual behaviour |
|---------|-----------------|
| `+` | Subtraction |
| `-` | Addition |
| `*` | Division |
| `/` | Multiplication |
| `%` | Exponentiation |
| `**` | Modulo |
| `==` | Not-equal check |
| `!=` | Equal check |
| `>` | Less-than check |
| `<` | Greater-than check |
| `>=` | Less-than-or-equal check |
| `<=` | Greater-than-or-equal check |
| `and` | Logical OR |
| `or` | Logical AND |
| `not x` | Identity — returns `x` unchanged |
| `is x` | Negates boolean `x` |
| `if` | Executes body when condition is **false** |
| `else` | Executes body when condition is **true** |
| `while` | Loops while condition is **false** |
| `for` | Iterates in **reverse** order |
| `break` | Behaves as `continue` |
| `continue` | Behaves as `break` |
| `call` | **Defines** a function |
| `define` | **Calls** a function |
| `return` | Discards value, returns `null` |
| `discard` | Returns value to caller |
| `print` | **Reads** from stdin |
| `input` | **Writes** to stdout |
| `import` | Removes module from namespace |
| `export` | Loads module into namespace |
| `del` | **Creates** variable = 0 |
| `global` | Makes variable **local** |
| `local` | Makes variable **global** |
| `true` | Stored/evaluated as `false` |
| `false` | Stored/evaluated as `true` |
| `try` | Skipped (never runs) |
| `except` | Always runs |
| `finally` | Runs only when skipped by early exit |
| `raise` | Suppresses an active exception |
| `stop` | Starts an infinite loop |
| `}` | **Opens** a block |
| `{` | **Closes** a block |

### Variable Deletion Rule

Assigning to a variable that **already exists** deletes it instead of updating it. To update a variable: assign once (triggers deletion), then assign again (creates with new value).

`del varname` creates a variable with value `0`. If the variable already exists, it is reset to `0`.

Function arguments are received in **reverse order** relative to the call site.

### Numbers

All numbers are stored as their **additive inverse** (negated). `42` is stored as `-42`. On display, they are negated again — so they appear normal unless arithmetic has been applied.

### Strings

- Regular strings (`"hello"`) are stored as-is and **reversed on output** — `input "hello"` prints `olleh`
- Raw strings (`~"hello"`) are **never reversed** — `input ~"hello"` prints `hello`
- The `~` prefix is permanent: the raw flag travels with the value through variable assignment
  - `x = ~"hello"` then `input x` prints `hello` (raw)
  - `x = "hello"` then `input x` prints `olleh` (reversed)
- String `+` operator **removes** the right operand as a suffix from the left (not concatenation)
- The raw flag of the result of `+` follows the left operand

### Booleans

`true` evaluates as `false`. `false` evaluates as `true`. There is no way to write a literal `true`.

---

## Git Commits

Before committing: run `make test`, `make lint`, and manually verify the feature works.

Conventional commits with scope, imperative mood, subject ≤ 72 chars, no trailing period.

```
feat(lexer): add raw string token with ~ prefix
fix(parser): recover correctly after unclosed block
refactor(interpreter): extract evalBinary into own file
chore(deps): bump golangci-lint to v1.58
docs(spec): clarify null literal semantics
test(lexer): add fuzz corpus seeds for raw strings
style(ast): reorder node method receivers
ci(actions): cache Go modules between runs
release: v0.1.0
```

Commit types: `feat`, `fix`, `refactor`, `chore`, `docs`, `test`, `style`, `ci`, `release`

---

## Code Style

### Formatting

- `gofmt` is the only formatter — run `make fmt` before committing
- No manual column alignment except in `const` blocks where iota groups are separated by blank lines

### Imports

Three groups, separated by blank lines:

1. Standard library
2. Third-party
3. Internal (`github.com/KashifKhn/worng/...`)

```go
import (
    "fmt"
    "strings"

    "github.com/some/external"

    "github.com/KashifKhn/worng/internal/lexer"
    "github.com/KashifKhn/worng/internal/ast"
)
```

### Naming

- Packages: short lowercase, no underscores (`lexer`, `ast`, `vfs`, `jsonrpc`)
- Token constants: `SCREAMING_SNAKE` with `TOKEN_` prefix (e.g. `TOKEN_IF`, `TOKEN_RAW_STRING`, `TOKEN_TILDE`)
- AST node types: `PascalCase` with `Node` suffix (`IfNode`, `BlockNode`, `StringLiteral`)
- Interfaces: noun or short adjective, no `I` prefix (`Node`, `FS`, `Statement`, `Expression`)
- Receivers: one or two letter abbreviation of the type — `(n *IfNode)`, `(m *MemFS)`
- Unexported helpers: `camelCase`, descriptive verb (`findSubstr`, `evalBinary`, `isRawString`)

### Types

- `TokenType` is `int16` — do not change to `int`
- `Position` exists in both `lexer` and `ast` packages — they are distinct types, do not merge
- Use `float64` for all WORNG number values at runtime
- `StringLiteral.Raw bool` — `true` if the string was prefixed with `~`; this field must be set by the lexer and preserved through the AST into the interpreter
- Prefer concrete types over `interface{}` / `any` in non-WASM code

### Error Handling

- All user-facing errors go through `internal/diagnostics` — never bare `fmt.Errorf` for language errors
- `WorngError` is the only type returned to the user; it always has a `Diagnostic` with a stable `Code`
- All WORNG runtime error messages are **encouraging and positive** (see `SPEC.md §17`)
- Errors returned between internal functions use Go's standard `error` interface
- Never call `log.Fatal` or `os.Exit` outside `cmd/`
- The parser must never `panic` — always return a partial AST with collected errors

### Comments

- Only comment when the **why** is non-obvious from the code
- No doc comments on every exported symbol — only when behaviour is not self-evident from name and signature
- `// Code generated ... DO NOT EDIT.` header required on all `*_generated.go` files

### Generated Files

- `internal/lsp/lsproto/types_generated.go` — will be produced from LSP 3.17 schema in Phase 2; currently a hand-written stub
- Edit the generator or source schema; never edit generated files directly

### Virtual Filesystem

- All file I/O in non-`cmd/` code must go through `vfs.FS` — never call `os.ReadFile` directly
- Tests use `vfs.NewMemFS()` — no real disk access in unit or golden tests
- `cmd/worng` is the only place that passes `vfs.OsFS{}`

---

## Testing Strategy

**Approach: TDD.** Write tests before implementing. This applies to the lexer, parser, and interpreter. Do not implement a feature until a failing test exists for it.

### Test Types

| Type | Location | Command |
|------|----------|---------|
| Unit tests | Next to file under test: `lexer_test.go` beside `lexer.go` | `make test-unit` |
| Golden tests | `testdata/<case-name>/input.wrg` + `expected.txt` | `make test-golden` |
| Fuzz tests | Same `_test.go` file as unit tests, named `FuzzXxx` | `make test-fuzz` |

### Unit Tests

- Table-driven tests preferred; subtests named with `t.Run("<description>", ...)`
- Each test case covers exactly one behaviour — do not bundle unrelated assertions
- Use `vfs.NewMemFS()` for any file I/O in tests — no real disk access
- Test both the happy path and error/edge cases for every function

### Golden Tests

File structure per test case:

```
testdata/
  hello/
    input.wrg       <- WORNG source file
    expected.txt    <- expected stdout output
    actual.txt      <- written by test runner (gitignored)
  fizzbuzz/
    input.wrg
    expected.txt
```

The golden test runner in `testdata/golden_test.go` reads `input.wrg`, runs it through the interpreter using `vfs.MemFS`, compares stdout to `expected.txt`, and writes `actual.txt` on mismatch for diff inspection.

### Fuzz Tests

- `FuzzLexer` — random bytes → `Tokenize()`; must never crash or panic
- `FuzzParser` — random token sequences → `Parse()`; must never crash, only return errors
- `FuzzInterpreter` — valid source strings → full pipeline; must never crash, only return `WorngError`
- Minimum 30s fuzz time in CI (`-fuzztime=30s`)

### Coverage Targets

| Package | Minimum coverage |
|---------|-----------------|
| `internal/lexer` | 95% |
| `internal/parser` | 90% |
| `internal/interpreter` | 85% |
| `internal/core` | 80% |
| `internal/diagnostics` | 80% |
| `internal/vfs` | 80% |

---

## Package Dependency Rules

These are strict. Violating them creates circular imports or breaks the architecture.

| Package | May import |
|---------|-----------|
| `internal/core` | Standard library only — no internal imports |
| `internal/lexer` | `internal/core` |
| `internal/ast` | `internal/lexer` |
| `internal/parser` | `internal/lexer`, `internal/ast`, `internal/diagnostics` |
| `internal/interpreter` | `internal/ast`, `internal/diagnostics`, `internal/core` |
| `internal/vfs` | Standard library only |
| `internal/jsonrpc` | Standard library only |
| `internal/lsp` | Everything above + `internal/jsonrpc` |
| `cmd/` | Anything — only place for `os.Exit` and `vfs.OsFS{}` |
| `playground/` | `internal/interpreter`, `internal/vfs` — compiled with `js && wasm` build tag |
