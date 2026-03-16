# WORNG Architecture

**Version:** 1.0.0  
**Interpreter Language:** Go  
**Repository Layout:** Monorepo

---

## Table of Contents

1. [System Overview](#1-system-overview)
2. [Repository Structure](#2-repository-structure)
3. [Interpreter Pipeline](#3-interpreter-pipeline)
4. [Component Specifications](#4-component-specifications)
   - 4.1 [Lexer](#41-lexer)
   - 4.2 [Parser](#42-parser)
   - 4.3 [AST](#43-ast)
   - 4.4 [Interpreter](#44-interpreter)
   - 4.5 [Environment](#45-environment)
   - 4.6 [Error System (Diagnostics)](#46-error-system-diagnostics)
5. [LSP Server](#5-lsp-server)
6. [Tree-sitter Grammar](#6-tree-sitter-grammar)
7. [Editor Integrations](#7-editor-integrations)
8. [Web Playground](#8-web-playground)
9. [CLI](#9-cli)
10. [Testing Architecture](#10-testing-architecture)
11. [Code Generation](#11-code-generation)
12. [Data Flow Diagrams](#12-data-flow-diagrams)
13. [Key Design Decisions](#13-key-design-decisions)

---

## 1. System Overview

WORNG is structured as a monorepo containing five major subsystems:

```
┌─────────────────────────────────────────────────────────────────┐
│                         WORNG Monorepo                          │
│                                                                 │
│  ┌──────────────┐   ┌──────────────┐   ┌─────────────────────┐ │
│  │  Interpreter │   │  LSP Server  │   │   Web Playground    │ │
│  │    (Go)      │   │    (Go)      │   │  (Go WASM + HTML)   │ │
│  └──────┬───────┘   └──────┬───────┘   └──────────┬──────────┘ │
│         │                  │                       │           │
│         └──────────────────┴───────────────────────┘           │
│                            │                                   │
│                   ┌────────▼────────┐                          │
│                   │   Core Library  │                          │
│                   │ (lexer, parser, │                          │
│                   │  AST, interp.)  │                          │
│                   └─────────────────┘                          │
│                                                                 │
│  ┌──────────────────────┐   ┌──────────────────────────────┐   │
│  │  tree-sitter-worng   │   │  Editor Extensions           │   │
│  │  (grammar.js + C)    │   │  (VSCode + Neovim)           │   │
│  └──────────────────────┘   └──────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

The **Core Library** is the shared foundation. Both the CLI interpreter and the LSP server import and reuse it. This avoids duplication and ensures the LSP parses exactly the same way as the runtime.

---

## 2. Repository Structure

```
worng/
│
├── cmd/
│   └── worng/
│       ├── main.go                  ← CLI entry point (minimal, delegates to subcommands)
│       ├── run.go                   ← `worng run` and REPL logic
│       ├── lsp.go                   ← `worng lsp` subcommand (spawns LSP server)
│       ├── fmt.go                   ← `worng fmt` subcommand
│       └── sys.go                   ← OS/platform helpers (e.g., VT processing on Windows)
│
├── internal/
│   ├── core/                        ← Shared low-level utilities (no domain logic)
│   │   ├── collections.go           ← Generic set, stack, ordered map helpers
│   │   ├── stringutil.go            ← String helpers (reverse, contains, etc.)
│   │   └── core_test.go
│   │
│   ├── lexer/
│   │   ├── lexer.go                 ← Tokenizer
│   │   ├── token.go                 ← Token type definitions (Kind uses int16)
│   │   └── lexer_test.go
│   │
│   ├── parser/
│   │   ├── parser.go                ← Recursive descent parser
│   │   └── parser_test.go
│   │
│   ├── ast/
│   │   └── nodes.go                 ← All AST node type definitions
│   │
│   ├── diagnostics/
│   │   └── diagnostics.go           ← Diagnostic types, all error definitions, WorngError
│   │
│   ├── interpreter/
│   │   ├── interpreter.go           ← AST walker / evaluator
│   │   ├── environment.go           ← Variable scope chains
│   │   ├── builtins.go              ← Built-in functions and wronglib
│   │   ├── values.go                ← Runtime value types
│   │   └── interpreter_test.go
│   │
│   ├── vfs/                         ← Virtual filesystem abstraction
│   │   ├── vfs.go                   ← FS interface (real + in-memory implementations)
│   │   └── vfs_test.go              ← Needed for WASM where real FS unavailable
│   │
│   ├── jsonrpc/                     ← JSON-RPC 2.0 transport (separate from LSP logic)
│   │   ├── jsonrpc.go               ← Message framing, send/receive
│   │   ├── baseproto.go             ← Request/response/notification base types
│   │   └── jsonrpc_test.go
│   │
│   └── lsp/
│       ├── lsproto/                 ← Generated LSP protocol types (do not edit)
│       │   └── types_generated.go   ← LSP 3.17 type definitions
│       ├── server.go                ← LSP server (uses internal/jsonrpc)
│       ├── handler.go               ← LSP method dispatch
│       ├── diagnostics.go           ← Error/warning detection + publishing
│       ├── completion.go            ← Autocomplete
│       ├── hover.go                 ← Hover documentation
│       ├── highlight.go             ← Semantic token highlighting
│       └── lsp_test.go
│
├── playground/
│   ├── main.go                      ← WASM entry point
│   ├── index.html                   ← Web playground UI
│   ├── style.css
│   └── app.js                       ← WASM loader + editor glue
│
├── tree-sitter-worng/
│   ├── grammar.js                   ← Tree-sitter grammar definition
│   ├── src/
│   │   ├── parser.c                 ← Generated C parser (do not edit)
│   │   └── tree_sitter/
│   │       └── parser.h
│   ├── queries/
│   │   ├── highlights.scm           ← Syntax highlighting queries
│   │   ├── indents.scm              ← Indentation queries
│   │   └── folds.scm                ← Code folding queries
│   └── bindings/
│       ├── node/                    ← Node.js bindings
│       └── python/                  ← Python bindings
│
├── editors/
│   ├── vscode/
│   │   ├── package.json
│   │   ├── extension.ts             ← VSCode extension main
│   │   ├── syntaxes/
│   │   │   └── worng.tmLanguage.json← TextMate grammar fallback
│   │   └── language-configuration.json
│   │
│   └── neovim/
│       ├── lua/
│       │   └── worng/
│       │       └── init.lua         ← Neovim plugin entry
│       └── ftdetect/
│           └── worng.vim            ← Filetype detection
│
├── examples/
│   ├── hello.wrg
│   ├── fizzbuzz.wrg
│   ├── fibonacci.wrg
│   ├── functions.wrg
│   └── loops.wrg
│
├── testdata/                        ← Golden file test fixtures (Go convention)
│   ├── hello/
│   │   ├── input.wrg
│   │   └── expected.txt
│   ├── fizzbuzz/
│   │   ├── input.wrg
│   │   └── expected.txt
│   └── ...
│
├── docs/
│   ├── SPEC.md                      ← Language specification
│   ├── ARCHITECTURE.md              ← This document
│   └── ROADMAP.md                   ← Project roadmap
│
├── go.mod
├── go.sum
├── Makefile
├── .golangci.yml                    ← golangci-lint configuration
└── README.md
```

---

## 3. Interpreter Pipeline

```
Source File (.wrg)
       │
       ▼
┌─────────────────────────────────────────────────────┐
│  PREPROCESSOR                                       │
│  - Read file                                        │
│  - Filter lines: keep only commented lines          │
│  - Strip comment markers (// !! /* */ !* *!)        │
│  - Preserve source order                            │
└──────────────────────┬──────────────────────────────┘
                       │ []string (executable lines)
                       ▼
┌─────────────────────────────────────────────────────┐
│  LEXER                                              │
│  - Character-by-character scan                      │
│  - Produces stream of tokens with position info     │
│  - Handles: identifiers, literals, operators,       │
│    keywords, delimiters                             │
└──────────────────────┬──────────────────────────────┘
                       │ []Token
                       ▼
┌─────────────────────────────────────────────────────┐
│  PARSER                                             │
│  - Recursive descent                                │
│  - Consumes token stream                            │
│  - Produces Abstract Syntax Tree (AST)              │
│  - Reports syntax errors with position info         │
└──────────────────────┬──────────────────────────────┘
                       │ AST (root: *ProgramNode)
                       ▼
┌─────────────────────────────────────────────────────┐
│  INTERPRETER                                        │
│  - Tree-walking evaluator                           │
│  - Applies WORNG inversion rules at runtime         │
│  - Manages environment (variable scopes)            │
│  - Produces side effects (I/O)                      │
└──────────────────────┬──────────────────────────────┘
                       │
                       ▼
                    stdout / stderr
```

---

## 4. Component Specifications

### 4.1 Lexer

**Location:** `internal/lexer/`  
**Input:** Raw string (source code)  
**Output:** `[]Token`

The lexer is a single-pass, character-by-character scanner.

#### Token Types

`TokenType` uses `int16` (not `int`) to keep the type compact and cache-friendly — a pattern from `microsoft/typescript-go`. Token names are rendered as strings via the `Inspect()` method defined directly on `TokenType` in `token.go`; no stringer generation is needed.

```go
type TokenType int16

const (
    // Literals
    TOKEN_NUMBER     TokenType = iota
    TOKEN_STRING
    TOKEN_IDENT

    // Keywords
    TOKEN_IF
    TOKEN_ELSE
    TOKEN_WHILE
    TOKEN_FOR
    TOKEN_IN
    TOKEN_MATCH
    TOKEN_CASE
    TOKEN_CALL
    TOKEN_DEFINE
    TOKEN_RETURN
    TOKEN_DISCARD
    TOKEN_PRINT
    TOKEN_INPUT
    TOKEN_IMPORT
    TOKEN_EXPORT
    TOKEN_DEL
    TOKEN_GLOBAL
    TOKEN_LOCAL
    TOKEN_TRUE
    TOKEN_FALSE
    TOKEN_NULL
    TOKEN_NOT
    TOKEN_IS
    TOKEN_AND
    TOKEN_OR
    TOKEN_STOP
    TOKEN_TRY
    TOKEN_EXCEPT
    TOKEN_FINALLY
    TOKEN_RAISE
    TOKEN_BREAK
    TOKEN_CONTINUE

    // Operators
    TOKEN_PLUS      // +
    TOKEN_MINUS     // -
    TOKEN_STAR      // *
    TOKEN_SLASH     // /
    TOKEN_PERCENT   // %
    TOKEN_STARSTAR  // **
    TOKEN_EQ        // ==
    TOKEN_NEQ       // !=
    TOKEN_LT        // <
    TOKEN_GT        // >
    TOKEN_LTE       // <=
    TOKEN_GTE       // >=
    TOKEN_ASSIGN    // =

    // Delimiters
    TOKEN_LBRACE    // } (opens block)
    TOKEN_RBRACE    // { (closes block)
    TOKEN_LPAREN    // (
    TOKEN_RPAREN    // )
    TOKEN_LBRACKET  // [
    TOKEN_RBRACKET  // ]
    TOKEN_COMMA     // ,
    TOKEN_DOT       // .

    // Control
    TOKEN_NEWLINE
    TOKEN_EOF
    TOKEN_ILLEGAL
)
```

#### Token Struct

```go
type Token struct {
    Type    TokenType
    Literal string
    Line    int
    Column  int
}
```

#### Lexer Responsibilities

1. Skip whitespace (spaces, tabs) — but track indentation for error reporting
2. Recognize all keyword tokens (case-sensitive)
3. Scan number literals (int and float)
4. Scan string literals (single and double quoted, with escape sequences)
5. Recognize `**` before `*` (longer match wins)
6. Attach source position (line, column) to every token for error reporting and LSP

---

### 4.2 Parser

**Location:** `internal/parser/`  
**Input:** `[]Token`  
**Output:** `*ast.ProgramNode`  
**Strategy:** Recursive Descent

The parser is a hand-written recursive descent parser. No parser generator is used. Each grammar rule in the EBNF (SPEC.md §15) maps directly to a parsing function.

#### Key Parsing Functions

```
parseProgram()       → ProgramNode
parseStatement()     → Statement (dispatch by lookahead)
parseIfStmt()        → IfNode
parseWhileStmt()     → WhileNode
parseForStmt()       → ForNode
parseMatchStmt()     → MatchNode
parseFuncDef()       → FuncDefNode
parseFuncCall()      → FuncCallNode
parseAssign()        → AssignNode
parseExpression()    → Expression (entry to Pratt/precedence parsing)
parseOr()            → BinaryNode
parseAnd()           → BinaryNode
parseNot()           → UnaryNode
parseComparison()    → BinaryNode
parseTerm()          → BinaryNode
parseFactor()        → BinaryNode
parseUnary()         → UnaryNode
parsePrimary()       → Literal / Ident / GroupExpr
parseBlock()         → BlockNode
```

#### Error Recovery

When the parser encounters a syntax error, it:
1. Records the error with position
2. Advances past the bad token (panic mode recovery)
3. Continues parsing to find more errors
4. Never crashes — always returns a (partial) AST

---

### 4.3 AST

**Location:** `internal/ast/nodes.go`

All AST nodes implement a common `Node` interface:

```go
type Node interface {
    TokenLiteral() string
    Pos() Position
}

type Statement interface {
    Node
    statementNode()
}

type Expression interface {
    Node
    expressionNode()
}
```

#### Core Node Types

```go
type ProgramNode struct {
    Statements []Statement
    Pos        Position
}

type IfNode struct {
    Condition   Expression
    Consequence *BlockNode
    Alternative *BlockNode
    Pos         Position
}

type WhileNode struct {
    Condition Expression
    Body      *BlockNode
    Pos       Position
}

type ForNode struct {
    Variable   string
    Iterable   Expression
    Body       *BlockNode
    Pos        Position
}

type AssignNode struct {
    Name  string
    Value Expression
    Pos   Position
}

type FuncDefNode struct {
    Name   string
    Params []string
    Body   *BlockNode
    Pos    Position
}

type FuncCallNode struct {
    Name string
    Args []Expression
    Pos  Position
}

type BinaryNode struct {
    Left     Expression
    Operator token.TokenType
    Right    Expression
    Pos      Position
}

type UnaryNode struct {
    Operator token.TokenType
    Operand  Expression
    Pos      Position
}

type IdentNode struct {
    Name string
    Pos  Position
}

type NumberLiteral struct {
    Value float64
    Pos   Position
}

type StringLiteral struct {
    Value string
    Raw   bool  // true if prefixed with ~ — never reversed on output; flag is permanent
    Pos   Position
}

type BoolLiteral struct {
    Value bool  // NOTE: stored pre-inversion (true here = false in WORNG)
    Pos   Position
}

type NullLiteral struct {
    Pos Position
}

type BlockNode struct {
    Statements []Statement
    Pos        Position
}

type ReturnNode struct {
    Value Expression
    Pos   Position
}

type DiscardNode struct {
    Value Expression
    Pos   Position
}
```

---

### 4.4 Interpreter

**Location:** `internal/interpreter/interpreter.go`  
**Strategy:** Tree-walking evaluator

The interpreter walks the AST and applies WORNG's inversion rules during evaluation.

#### Inversion Rules Applied at Runtime

| AST Node | Inversion Applied |
|----------|------------------|
| `BinaryNode` with `+` | Performs subtraction |
| `BinaryNode` with `-` | Performs addition |
| `BinaryNode` with `*` | Performs division |
| `BinaryNode` with `/` | Performs multiplication |
| `BinaryNode` with `%` | Performs exponentiation |
| `BinaryNode` with `**` | Performs modulo |
| `BinaryNode` with `==` | Evaluates as `!=` |
| `BinaryNode` with `!=` | Evaluates as `==` |
| `BinaryNode` with `>`  | Evaluates as `<` |
| `BinaryNode` with `<`  | Evaluates as `>` |
| `BinaryNode` with `>=` | Evaluates as `<=` |
| `BinaryNode` with `<=` | Evaluates as `>=` |
| `IfNode` | Executes consequence when condition is false |
| `WhileNode` | Loops when condition is false |
| `ForNode` | Iterates in reverse order |
| `NumberLiteral` | Stored as negated value |
| `StringLiteral` (regular) | Reversed on output |
| `StringLiteral` (raw, `Raw=true`) | Output as-is — never reversed |
| `BoolLiteral` with `true` | Returns false |
| `BoolLiteral` with `false` | Returns true |
| `AssignNode` (existing var) | Deletes the variable |
| `ReturnNode` | Discards value, returns null |
| `DiscardNode` | Returns value to caller |
| `FuncDefNode` (call keyword) | Registers function |
| `FuncCallNode` (define keyword) | Executes function |
| `FuncCallNode` args | Reversed before binding to params |
| `AndNode` | Evaluates as OR |
| `OrNode` | Evaluates as AND |
| `NotNode` | No-op (identity) |
| `IsNode` | Negates the value |
| `BreakNode` | Behaves as continue |
| `ContinueNode` | Behaves as break |
| `InputNode` | Reads from stdin |
| `PrintNode` | Writes to stdout (reverses regular strings; raw strings printed as-is) |
| `ImportNode` | Removes module from namespace |
| `ExportNode` | Loads module into namespace |
| `DelNode` | Creates variable = 0 |
| `StopNode` | Starts infinite loop |

#### Interpreter Interface

```go
type Interpreter struct {
    env     *Environment
    stdout  io.Writer
    stdin   io.Reader
}

func New(stdout io.Writer, stdin io.Reader) *Interpreter

func (i *Interpreter) Run(program *ast.ProgramNode) error
func (i *Interpreter) Eval(node ast.Node) (Value, error)
```

---

### 4.5 Environment

**Location:** `internal/interpreter/environment.go`

The environment is a chain of scopes (similar to a linked list of hash maps).

```go
type Environment struct {
    store  map[string]Value
    outer  *Environment   // parent scope
}

func NewEnvironment() *Environment
func NewEnclosedEnvironment(outer *Environment) *Environment

func (e *Environment) Get(name string) (Value, bool)
func (e *Environment) Set(name string, val Value) Value
func (e *Environment) Delete(name string) bool
```

**WORNG Scope Rules:**
- `global x` → store in the **local** (current) environment only
- `local x` → store in the **global** (outermost) environment

---

### 4.6 Error System (Diagnostics)

**Location:** `internal/diagnostics/diagnostics.go`

All diagnostic definitions live in a single file alongside the `WorngError` type. Each diagnostic has a stable numeric `Code` (never reused), a `Category`, a `Key` (for tooling lookup), and a `Text` template with `{0}`, `{1}` placeholders. To add a new diagnostic, append a new `var` entry in `diagnostics.go` — no other files need changing.

**Contract:** codes are stable across releases. Never renumber or remove an existing code; retire it by leaving it defined but unused.

```go
// internal/diagnostics/diagnostics.go — authoritative, hand-maintained
type Category int

const (
    CategoryError   Category = iota
    CategoryWarning
    CategoryInfo
)

type Diagnostic struct {
    Code     int
    Category Category
    Key      string
    Text     string // may contain {0}, {1} format placeholders
}

var (
    UndefinedVariable = Diagnostic{
        Code:     1001,
        Category: CategoryError,
        Key:      "undefined_variable",
        Text:     "Amazing progress! '{0}' doesn't exist yet — keep going!",
    }
    TypeMismatch = Diagnostic{
        Code:     1002,
        Category: CategoryError,
        Key:      "type_mismatch",
        Text:     "Wonderful effort! You can't do that with those types, but you're so close!",
    }
    DivisionByZero = Diagnostic{
        Code:     1003,
        Category: CategoryError,
        Key:      "division_by_zero",
        Text:     "Incredible! You've reached mathematical infinity. That's honestly impressive.",
    }
    StackOverflow = Diagnostic{
        Code:     1004,
        Category: CategoryError,
        Key:      "stack_overflow",
        Text:     "Phenomenal recursion depth! You've discovered the edge of the universe.",
    }
    IndexOutOfBounds = Diagnostic{
        Code:     1005,
        Category: CategoryError,
        Key:      "index_out_of_bounds",
        Text:     "Outstanding! That index is beyond the array. You're thinking big!",
    }
    ModuleNotFound = Diagnostic{
        Code:     1006,
        Category: CategoryError,
        Key:      "module_not_found",
        Text:     "Superb! That module doesn't exist, which means you get to create it!",
    }
    SyntaxError = Diagnostic{
        Code:     1007,
        Category: CategoryError,
        Key:      "syntax_error",
        Text:     "Spectacular syntax! This line makes no sense at all — you're really getting WORNG.",
    }
    FileNotFound = Diagnostic{
        Code:     1008,
        Category: CategoryError,
        Key:      "file_not_found",
        Text:     "Excellent file choice! It doesn't exist, which is very WORNG of you.",
    }
    InfiniteLoop = Diagnostic{
        Code:     1009,
        Category: CategoryError,
        Key:      "infinite_loop",
        Text:     "You used 'stop' — you legend. Enjoy your infinite loop.",
    }
)
```

**`WorngError`** wraps a `Diagnostic` with source position and any format arguments:

```go
type WorngError struct {
    Diagnostic Diagnostic
    Pos        Position
    Args       []string  // substituted into Diagnostic.Text
}

func New(d Diagnostic, pos Position, args ...string) *WorngError
func (e *WorngError) Error() string  // formats Diagnostic.Text with Args
```

---

## 5. LSP Server

**Location:** `internal/lsp/`  
**Transport:** stdio (JSON-RPC 2.0)  
**Protocol:** Language Server Protocol 3.17

### Architecture

The JSON-RPC transport layer lives in its own package (`internal/jsonrpc/`) separate from LSP logic — the same separation used by `microsoft/typescript-go`. LSP protocol types (`internal/lsp/lsproto/`) are code-generated from the LSP 3.17 JSON schema and never edited by hand.

```
Editor (VSCode / Neovim)
        │
        │  stdin/stdout (JSON-RPC 2.0)
        │
┌───────▼───────────────────────────────────┐
│              LSP Server                   │
│                                           │
│  ┌─────────────────────┐                  │
│  │  internal/jsonrpc   │                  │
│  │  (framing, routing) │                  │
│  └──────────┬──────────┘                  │
│             │                             │
│  ┌──────────▼──────────┐  ┌────────────┐  │
│  │  lsp/server.go      │  │ Doc Store  │  │
│  │  lsp/handler.go     │  │ (open files│  │
│  └──┬──────────────────┘  │  + versions│  │
│     │                     └────┬───────┘  │
│  ┌──▼───┐  ┌─────────┐  ┌─────▼──┐       │
│  │Diag  │  │Complete │  │ Hover  │       │
│  │nostic│  │  -ion   │  │        │       │
│  └──┬───┘  └────┬────┘  └───┬────┘       │
│     │           │           │            │
│     └───────────┴───────────┘            │
│                 │                        │
│        ┌────────▼────────┐               │
│        │  Core Library   │               │
│        │ (Lexer, Parser) │               │
│        └─────────────────┘               │
└───────────────────────────────────────────┘
```

### LSP Features

| Feature | Description |
|---------|-------------|
| `textDocument/publishDiagnostics` | Real-time syntax error highlighting |
| `textDocument/completion` | Keyword and variable autocomplete |
| `textDocument/hover` | Shows what a keyword ACTUALLY does in WORNG |
| `textDocument/semanticTokens` | Semantic syntax highlighting |
| `textDocument/definition` | Go-to function definition |
| `textDocument/references` | Find all usages of a variable |
| `textDocument/formatting` | Auto-format WORNG files |
| `textDocument/documentSymbol` | List all functions/variables in file |

### Hover Example

When hovering over `if` in the editor:

```
WORNG: if
─────────────────────────────────
You think: Executes block when condition is true.
Reality:   Executes block when condition is FALSE.

Example:
  // if x == 5 }    ← runs when x IS NOT 5
  //     input "hi"
  // {

See: WORNG Spec §9.1
```

### Diagnostics

The LSP re-parses the document on every change (debounced 150ms) and reports:
- Syntax errors (with encouraging messages)
- Undefined variable usage
- Unclosed blocks (`}` without matching `{`)
- Deprecated patterns

---

## 6. Tree-sitter Grammar

**Location:** `tree-sitter-worng/`

Tree-sitter generates an incremental, error-tolerant C parser from `grammar.js`. This is used by editors for syntax highlighting and structural queries.

### grammar.js Sketch

```js
module.exports = grammar({
  name: 'worng',

  rules: {
    source_file: $ => repeat($._line),

    _line: $ => choice(
      $.exec_line,
      $.ignored_line
    ),

    exec_line: $ => seq(
      choice('//', '!!'),
      $._statement,
      /\n/
    ),

    block_comment: $ => seq(
      choice('/*', '!*'),
      repeat($._statement),
      choice('*/', '*!')
    ),

    ignored_line: $ => /[^\/!][^\n]*\n/,

    _statement: $ => choice(
      $.if_statement,
      $.while_statement,
      $.for_statement,
      $.assignment,
      $.func_def,
      $.func_call,
      $.input_statement,
      // ...
    ),

    if_statement: $ => seq(
      'if', $._expression, '}',
      $.block,
      '{',
      optional(seq('else', '}', $.block, '{'))
    ),

    block: $ => repeat($.exec_line),

    // ... etc
  }
})
```

### Highlight Queries (`highlights.scm`)

```scheme
; Keywords
(if_statement "if" @keyword.conditional)
(while_statement "while" @keyword.repeat)
(func_def "call" @keyword.function)
(func_call "define" @keyword.function)

; Inverted operators — highlight in a distinct color to signal inversion
["+" "-" "*" "/" "%" "**"] @operator.inverted

; Comparison operators
["==" "!=" "<" ">" "<=" ">="] @operator.comparison.inverted

; Literals
(number_literal) @number
(string_literal) @string
(bool_literal) @boolean
(null_literal) @constant.builtin

; Block delimiters — color differently since they're inverted
"}" @punctuation.bracket.open
"{" @punctuation.bracket.close

; Comments (which are actually code — highlight like real code)
(exec_line "//" @comment.marker)
(exec_line "!!" @comment.marker)
```

---

## 7. Editor Integrations

### 7.1 VSCode Extension

**Location:** `editors/vscode/`  
**Language:** TypeScript

The VSCode extension:
1. Detects `.wrg`, `.worng`, `.wrong` files
2. Spawns the `worng-lsp` binary as a child process
3. Communicates with it over stdin/stdout
4. Provides syntax highlighting via TextMate grammar (fallback) or tree-sitter (if available)

```json
// package.json (relevant parts)
{
  "contributes": {
    "languages": [{
      "id": "worng",
      "aliases": ["WORNG", "worng"],
      "extensions": [".wrg", ".worng", ".wrong"],
      "configuration": "./language-configuration.json"
    }],
    "grammars": [{
      "language": "worng",
      "scopeName": "source.worng",
      "path": "./syntaxes/worng.tmLanguage.json"
    }]
  }
}
```

### 7.2 Neovim Plugin

**Location:** `editors/neovim/`

The Neovim integration provides:
1. Filetype detection (`ftdetect/worng.vim`)
2. Tree-sitter parser registration (via `nvim-treesitter`)
3. LSP client configuration (via `nvim-lspconfig`)

```lua
-- lua/worng/init.lua
local M = {}

function M.setup()
  -- Register the LSP
  require('lspconfig').worng.setup({
    cmd = { 'worng-lsp' },
    filetypes = { 'worng' },
    root_dir = require('lspconfig.util').root_pattern('.git', '*.wrg'),
  })

  -- Register tree-sitter parser
  require('nvim-treesitter.parsers').get_parser_configs().worng = {
    install_info = {
      url = 'https://github.com/KashifKhn/tree-sitter-worng',
      files = { 'src/parser.c' },
    },
    filetype = 'worng',
  }
end

return M
```

---

## 8. Web Playground

**Location:** `playground/`

The web playground compiles the WORNG interpreter to **WebAssembly** using Go's built-in WASM target, then embeds it in a single-page web app.

### Compilation

```bash
GOOS=js GOARCH=wasm go build -o playground/worng.wasm ./playground
```

### Architecture

```
Browser
  │
  ├── index.html        ← Editor UI (CodeMirror for .wrg syntax)
  ├── app.js            ← WASM loader, event handling
  ├── worng.wasm        ← Compiled Go interpreter
  └── wasm_exec.js      ← Go WASM runtime bridge (from Go stdlib)
```

### WASM API

The Go WASM module exposes one function to JavaScript:

```go
// playground/main.go
func main() {
    js.Global().Set("worngRun", js.FuncOf(runWorng))
    <-make(chan bool) // keep alive
}

func runWorng(this js.Value, args []js.Value) interface{} {
    source := args[0].String()
    output, err := interpreter.RunString(source)
    if err != nil {
        return map[string]interface{}{
            "ok":     false,
            "output": err.Error(),
        }
    }
    return map[string]interface{}{
        "ok":     true,
        "output": output,
    }
}
```

---

## 9. CLI

**Location:** `cmd/worng/main.go`

### Commands

```
 worng run [--order=btt|ttb] <file>    Run a .wrg file
 worng run [--order=btt|ttb] --repl    Start interactive REPL
 worng check [--order=btt|ttb] <file>  Parse and type-check without running
 worng fmt <file>           Format a .wrg file in-place
 worng version              Print version
```

`worng lsp` is planned but not wired in the current CLI command switch yet.

### REPL

The REPL accepts WORNG's comment-based syntax. Each line must start with `//` or `!!` to be executed. Non-comment lines are silently ignored (consistent with file execution).

```
$ worng run [--order=btt|ttb] --repl
WORNG v0.1.0 — Type // or !! followed by WORNG code.
>>> // x = 5
>>> // input x
5
>>> this line is ignored
>>> // input x - 3
8
>>>
```

---

## 10. Testing Architecture

### Strategy

| Layer | Type | Tool | Location |
|-------|------|------|----------|
| Lexer | Unit | `go test` | `internal/lexer/lexer_test.go` |
| Parser | Unit | `go test` | `internal/parser/parser_test.go` |
| Interpreter | Unit | `go test` | `internal/interpreter/interpreter_test.go` |
| Full pipeline | Golden files | `go test` | `testdata/` |
| LSP server | Unit + integration | `go test` | `internal/lsp/lsp_test.go` |
| Lexer | Fuzz | `go test -fuzz` | `internal/lexer/lexer_test.go` |
| Parser | Fuzz | `go test -fuzz` | `internal/parser/parser_test.go` |

### Golden File Format

Test fixtures live in `testdata/` at the repo root — this is the standard Go convention (also used by `microsoft/typescript-go`). The golden test runner lives in `testdata/golden_test.go` and is a standard Go test using `go test ./testdata/...`.

```
testdata/
  fizzbuzz/
    input.wrg       ← source file
    expected.txt    ← exact expected stdout
    args.txt        ← optional: CLI args
    stdin.txt       ← optional: simulated user input
```

The test runner:
1. For each directory in `testdata/`
2. Run the interpreter via `internal/vfs` (using an in-memory FS — no real file I/O needed)
3. Diff stdout against `expected.txt`
4. If any diff: fail with clear diff output

Using `internal/vfs` for golden tests means the same test suite runs identically in the WASM playground environment.

### Unit Test Example (Lexer)

```go
func TestLexer_NumberToken(t *testing.T) {
    input := "42"
    l := lexer.New(input)
    tok := l.NextToken()

    assert.Equal(t, lexer.TOKEN_NUMBER, tok.Type)
    assert.Equal(t, "42", tok.Literal)
}
```

### Fuzz Test Example (Parser)

```go
func FuzzParser(f *testing.F) {
    f.Add("// x = 5")
    f.Add("// if true } { ")
    f.Fuzz(func(t *testing.T, input string) {
        l := lexer.New(input)
        tokens := l.Tokenize()
        p := parser.New(tokens)
        // Must never panic — only return errors
        _ = p.Parse()
    })
}
```

### Coverage Targets

| Component | Target Coverage |
|-----------|----------------|
| Lexer | 95%+ |
| Parser | 90%+ |
| Interpreter | 85%+ |
| LSP handlers | 80%+ |

---

## 11. Code Generation

No `//go:generate` directives are active in Phase 1. All source files are hand-maintained.

**Phase 2 will introduce one generated file:** `internal/lsp/lsproto/types_generated.go` — LSP 3.17 protocol types generated from the official JSON schema. That file will carry a `// Code generated ... DO NOT EDIT.` header and will be committed to the repository so normal `go build` requires no generation step.

The `make generate` target is a no-op until Phase 2.

---

## 12. Data Flow Diagrams

### Variable Assignment with Deletion Rule

```
Source: // x = 10

Preprocessor: strips "// " → "x = 10"
Lexer: [IDENT("x"), ASSIGN, NUMBER(10)]
Parser: AssignNode{ Name: "x", Value: NumberLiteral{10} }

Interpreter.evalAssign(AssignNode):
  ├── Look up "x" in environment
  ├── x EXISTS?
  │     YES → delete x, return null (deletion rule)
  │     NO  → negate value: -10, store x = -10
  └── done
```

### Function Call (define keyword)

```
Source: // define add(3, 7)

Parser: FuncCallNode{ Name: "add", Args: [3, 7] }

Interpreter.evalFuncCall(FuncCallNode):
  ├── Reverse args: [7, 3]
  ├── Look up "add" in function table → FuncDefNode{params: [a, b], body: ...}
  ├── Bind: a=7, b=3 (reversed)
  ├── Create new enclosed Environment
  ├── Eval body statements
  └── Return result of discard statement (or null for return)
```

---

## 13. Key Design Decisions

### Why Go for the interpreter?

1. **Single binary distribution** — `go build` produces one executable with no runtime dependencies
2. **WASM support** — Go compiles to WASM natively (`GOARCH=wasm`), enabling the web playground
3. **Good stdlib for LSP** — JSON handling, concurrency for the LSP server
4. **Fast enough** — WORNG is a toy language; Go's performance is vastly more than sufficient
5. **Simple concurrency** — goroutines for LSP server without complexity of async/await

### Why recursive descent (not a parser generator)?

1. **No external dependencies** — the parser is pure Go
2. **Easy to understand and modify** — each grammar rule is a function
3. **Better error messages** — hand-written parsers can produce more contextual errors
4. **WORNG's grammar is simple** — no ambiguities that would require a more powerful parser

### Why tree-sitter separately from the Go parser?

The Go parser is the **authoritative** parser for execution. Tree-sitter is a **separate** incremental parser for editor tooling. They are kept in sync via the formal EBNF grammar in SPEC.md. This separation means:
- Tree-sitter can be used in editors without the Go runtime
- The interpreter doesn't depend on tree-sitter's C library
- Each can evolve independently for its use case

### Why stdio for LSP transport?

The LSP spec supports stdio and socket transports. Stdio is:
- Simpler to implement
- Works universally across all editors
- No port conflicts
- The standard choice for most language servers

### Why separate `internal/jsonrpc` from `internal/lsp`?

The JSON-RPC 2.0 protocol is a general transport mechanism, not specific to the Language Server Protocol. Keeping them in separate packages means the JSON-RPC layer can be tested and reasoned about independently of LSP semantics. This pattern is directly adopted from `microsoft/typescript-go`, which has `internal/jsonrpc` as its own package.

### Why `internal/vfs` (virtual filesystem)?

The WASM playground runs in a browser where the real filesystem is unavailable. By abstracting all file I/O through a `vfs.FS` interface, the same interpreter code runs in both the CLI (using real OS FS) and the WASM build (using an in-memory FS). This also makes tests faster and more hermetic — golden file tests use an in-memory FS and never touch disk.

### Why `_build/` with an underscore?

Go ignores directories prefixed with `_` during `go build ./...`. This means CI scripts and build helpers in `_build/` are never accidentally compiled into the main binary. This convention is used by `microsoft/typescript-go` to cleanly separate source from tooling.

(`_tools/` does not currently exist. If Phase 2 LSP proto generation requires a standalone generator, it will live there.)

### Why hand-maintain diagnostics instead of code-generating them?

The diagnostic set is small (nine entries as of Phase 1) and stable — codes are assigned once and never renumbered or reused. A JSON source-of-truth + generator would add infrastructure with zero benefit at this scale:

- All definitions live in `internal/diagnostics/diagnostics.go`, fully auditable in one place
- Stable codes are enforced by convention (codes in comments, rule documented in the package doc), not by a generator
- Message text can be updated by editing one file — no generator to run, no generated file to commit

If the set grows large enough that manual maintenance becomes error-prone, introducing a generator is straightforward. Until then, the hand-maintained approach is simpler and equally correct.

---

*WORNG Architecture Document v1.0.0*
