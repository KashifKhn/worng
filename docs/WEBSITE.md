# WORNG Documentation Website Plan

**Version:** 1.0.0  
**Status:** Pre-build planning  
**Site framework:** VitePress  
**Deployment:** GitHub Pages via CI  

> "The docs should be as wrong as the language. But readable."

---

## Table of Contents

1. [Overview](#1-overview)
2. [Stack and Tooling](#2-stack-and-tooling)
3. [Repository Layout](#3-repository-layout)
4. [Site Structure and Pages](#4-site-structure-and-pages)
5. [Visual Design](#5-visual-design)
6. [Playground Integration](#6-playground-integration)
7. [CI / Deployment](#7-ci--deployment)
8. [Phase-by-Phase Doc Sync Plan](#8-phase-by-phase-doc-sync-plan)
   - [Phase 0 — Foundation](#phase-0--foundation)
   - [Phase 1 — Core Interpreter](#phase-1--core-interpreter)
   - [Phase 2 — LSP Server](#phase-2--lsp-server)
   - [Phase 3 — Editor Integrations](#phase-3--editor-integrations)
   - [Phase 4 — Web Playground](#phase-4--web-playground)
   - [Phase 5 — Polish and Publish](#phase-5--polish-and-publish)
9. [Page Content Specifications](#9-page-content-specifications)
10. [VitePress Configuration Reference](#10-vitepress-configuration-reference)

---

## 1. Overview

The WORNG documentation site is the single source of truth for anyone wanting to learn, use, or contribute to the WORNG language. It covers:

- A full language reference (every keyword, operator, construct)
- A getting-started guide (install → first program → run it)
- Annotated examples gallery
- Interactive browser playground (WASM-powered)
- Architecture and contributor docs
- A live roadmap

The site must stay **in sync with the code**. Every time a phase of the interpreter, LSP, or editor integration ships, the corresponding doc pages are written or updated before that phase is considered done. This file tracks exactly what needs to be written at each step.

---

## 2. Stack and Tooling

| Concern | Tool |
|---------|------|
| Framework | [VitePress](https://vitepress.dev) — Vue-powered static site, Markdown-native |
| Language | Vue 3 + TypeScript (for custom components) |
| Playground UI | Custom Vue component — `WrongPlayground.vue` |
| Runtime in browser | Go compiled to WASM (`playground/worng.wasm`) |
| Code editor in playground | Monaco Editor (same as VSCode) |
| Deployment | GitHub Pages — `gh-pages` branch |
| CI | GitHub Actions — `.github/workflows/docs.yml` |
| Node version | ≥ 18 (LTS) |
| Package manager | `npm` |

### Why VitePress

- Markdown files are the source; no MDX complexity
- Vue components can be embedded directly in any `.md` page — perfect for the playground
- Default theme is clean and professional out of the box
- Fast Vite-based dev server (`npm run dev`)
- Algolia DocSearch integration available when the site is public
- Closest match to the Go, Rust, and TypeScript doc sites in simplicity and polish

---

## 3. Repository Layout

The site lives inside `docs/` in the existing repo. VitePress uses `docs/.vitepress/` as its config directory. All content pages are Markdown files directly in `docs/` or its subdirectories.

```
docs/
├── .vitepress/
│   ├── config.ts                    ← nav, sidebar, theme, SEO config
│   └── theme/
│       ├── index.ts                 ← extend default theme, register components
│       ├── style.css                ← WORNG colour palette + typography overrides
│       └── components/
│           └── WrongPlayground.vue  ← live WASM playground component
│
├── public/
│   ├── logo.svg                     ← WORNG logo (inverted W mark)
│   └── worng.wasm                   ← compiled WASM binary (placeholder until Phase 4)
│
├── index.md                         ← home page: hero, features, quick-start CTA
│
├── guide/
│   └── getting-started.md           ← install CLI, write first .wrg, run it
│
├── language/
│   ├── overview.md                  ← what WORNG is, the 4 design principles
│   ├── execution-model.md           ← bottom-to-top, comment/code rule
│   ├── data-types.md                ← numbers, strings, booleans, null
│   ├── operators.md                 ← all inverted arithmetic, comparison, logical
│   ├── control-flow.md              ← if/else, while, for, break/continue, match
│   ├── variables.md                 ← assignment, deletion rule, del, scope
│   ← functions.md                  ← call/define, params reversed, return/discard
│   ├── io.md                        ← input/print, inputln/println
│   ├── error-handling.md            ← try/except/finally/raise
│   ├── modules.md                   ← import/export, wronglib stdlib
│   ├── reserved-words.md            ← full keyword reference table
│   └── grammar.md                   ← EBNF grammar (verbatim from SPEC.md §15)
│
├── examples.md                      ← annotated examples gallery
├── playground.md                    ← full-page playground (embeds WrongPlayground.vue)
├── architecture.md                  ← system architecture (for contributors)
└── roadmap.md                       ← phase tracker (mirrors ROADMAP.md)
```

> `SPEC.md`, `ARCHITECTURE.md`, and `ROADMAP.md` remain as raw documents in `docs/`. The VitePress pages are separate, human-friendly versions of that content — they are not auto-generated from those files.

---

## 4. Site Structure and Pages

### Navigation (top bar)

```
Logo + "WORNG"    Guide    Language    Examples    Playground    GitHub (icon)
```

### Sidebar — Guide

```
Guide
  Getting Started
```

### Sidebar — Language Reference

```
Language Reference
  Overview
  Execution Model
  Data Types
  Operators
  Control Flow
  Variables
  Functions
  Input & Output
  Error Handling
  Modules
  Reserved Words
  Grammar (EBNF)
```

### Sidebar — Contribute / Internals

```
Internals
  Architecture
  Roadmap
```

---

## 5. Visual Design

**Personality:** Professional but playful. The docs must be readable and trustworthy. WORNG's weirdness comes through the content, not the layout.

### Colour Palette

| Token | Hex | Usage |
|-------|-----|-------|
| Brand primary | `#E84545` | Headings, links, hero accent, CTA buttons |
| Brand secondary | `#2B2D42` | Dark backgrounds, nav bar |
| Brand accent | `#F5A623` | Inline code, operator highlights, warnings |
| Surface | `#FAFAFA` | Page background (light mode) |
| Surface dark | `#1A1B26` | Page background (dark mode) |
| Text primary | `#1C1C1E` | Body text (light) |
| Text primary dark | `#E2E2E8` | Body text (dark) |
| Code block bg | `#282C34` | Shiki code blocks |

### Typography

- Body: system-ui stack (same as VitePress default)
- Code: `JetBrains Mono`, fallback to `monospace`
- All code examples use WORNG-aware syntax highlighting (custom Shiki grammar once tree-sitter grammar is available in Phase 3; plain text highlighting before that)

### Custom Elements

- `::before "WORNG"` callout boxes — styled like Rust's warning boxes but with WORNG's encouraging tone
- Inverted brace pairs `}` / `{` are highlighted in the block delimiter colour throughout the docs
- Error messages shown in examples use the encouraging format (no red scary boxes — soft amber)

---

## 6. Playground Integration

The playground page (`/playground`) embeds `WrongPlayground.vue` full-width.

### Component Layout

```
┌─────────────────────────────────────────────────────┐
│  WORNG Playground                    [Examples ▼]   │
├──────────────────────────┬──────────────────────────┤
│                          │                          │
│   Monaco Editor          │   Output                 │
│   (.wrg syntax)          │   (stdout / errors)      │
│                          │                          │
│                          │                          │
├──────────────────────────┴──────────────────────────┤
│  [Run ▶]  [Clear]  [Share 🔗]                       │
└─────────────────────────────────────────────────────┘
```

### Behaviour

| Action | Result |
|--------|--------|
| Click Run | Calls `worngRun(source)` from WASM; output appears in right panel |
| Click Clear | Clears editor and output |
| Click Share | Encodes editor content in URL fragment (`#code=<base64>`) |
| Select example | Loads preset program into editor |
| Keyboard shortcut | `Ctrl+Enter` / `Cmd+Enter` runs the program |

### Before Phase 4 (WASM not yet built)

The component renders the full UI but the Run button shows:

> "The playground is coming in Phase 4. For now, install the CLI and run locally."

The editor is still functional (Monaco loads, syntax highlight works after Phase 3).

---

## 7. CI / Deployment

### Workflow file: `.github/workflows/docs.yml`

```yaml
name: Docs

on:
  push:
    branches: [main]
    paths:
      - 'docs/**'
      - '.github/workflows/docs.yml'

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: npm
          cache-dependency-path: docs/package-lock.json

      - name: Install dependencies
        run: npm ci
        working-directory: docs

      - name: Build
        run: npm run build
        working-directory: docs

      - name: Deploy to GitHub Pages
        uses: peaceiris/actions-gh-pages@v4
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          publish_dir: docs/.vitepress/dist
```

### `docs/package.json` (minimal)

```json
{
  "name": "worng-docs",
  "private": true,
  "scripts": {
    "dev": "vitepress dev",
    "build": "vitepress build",
    "preview": "vitepress preview"
  },
  "devDependencies": {
    "vitepress": "^1.3.0",
    "@monaco-editor/loader": "^1.4.0"
  }
}
```

---

## 8. Phase-by-Phase Doc Sync Plan

This is the core of this document. For every code phase, there is a checklist of doc pages that must be written or updated **before that phase is marked done**. Doc work is part of the Definition of Done.

---

### Phase 0 — Foundation

**Code milestone:** Repository scaffolded, `go build ./...` clean.  
**Doc goal:** Site scaffolding exists. Placeholder pages in place. No content yet — but the site builds and deploys.

#### Checklist

- [ ] Create `docs/package.json` with VitePress dependency
- [ ] Create `docs/.vitepress/config.ts` — nav, sidebar skeleton, site title, description
- [ ] Create `docs/.vitepress/theme/index.ts` — extend default theme
- [ ] Create `docs/.vitepress/theme/style.css` — WORNG colour palette only (no content-specific styles yet)
- [ ] Create `docs/public/logo.svg` — WORNG logo (simple text-based or inverted-W mark)
- [ ] Create `docs/index.md` — home page with hero (placeholder tagline, "coming soon" body)
- [ ] Create all content page stubs (`guide/`, `language/`, `examples.md`, `playground.md`, `architecture.md`, `roadmap.md`) — each with just a `# Title` heading and a "This page is under construction" note
- [ ] Create `docs/public/worng.wasm` — empty placeholder file so the build does not 404
- [ ] Create `.github/workflows/docs.yml` — CI that builds and deploys to GitHub Pages
- [ ] Verify: `npm run build` inside `docs/` succeeds with zero errors
- [ ] Verify: site deploys to GitHub Pages and is accessible at `https://KashifKhn.github.io/worng/`

---

### Phase 1 — Core Interpreter

**Code milestone:** `worng run` works. All language features implemented. All golden tests pass.  
**Doc goal:** The full language reference is written. Anyone reading the docs can understand every feature of WORNG. The getting-started guide takes a reader from zero to running their first program.

#### After Phase 1.1 (Lexer) + 1.2 (Preprocessor)

- [ ] **`language/execution-model.md`** — write in full:
  - Bottom-to-top execution with a clear annotated example
  - The comment/code rule (only `//`, `!!`, `/* */`, `!* *!` lines execute)
  - Single-line vs block comment formats with examples
  - No-nesting rule for block comments
  - Visual diagram: source file → preprocessor → executable lines (reversed) → interpreter

#### After Phase 1.3 (AST) + 1.4 (Parser)

- [ ] **`language/overview.md`** — write in full:
  - What WORNG is (2-3 paragraphs)
  - The 4 design principles with brief explanation of each
  - "WORNG is not a joke" section
  - Link to every language section in the sidebar
- [ ] **`language/grammar.md`** — copy EBNF grammar from `SPEC.md §15` verbatim, add explanatory notes for each rule

#### After Phase 1.5 (Values) + 1.6 (Environment)

- [ ] **`language/data-types.md`** — write in full:
  - Numbers: internal negation, write `42` store `-42`, display negation chain
  - Strings: stored as-is, reversed on output by default — with `input "hello"` → `olleh` example
  - Raw strings: `~"hello"` — never reversed; raw flag travels with the value permanently
  - Side-by-side comparison: `input "hello"` vs `input ~"hello"`, `x = "hello"; input x` vs `x = ~"hello"; input x`
  - Booleans: `true` is `false`, `false` is `true` — table + example
  - Null: the one honest literal — explanation and rationale
  - Type coercion: WORNG does not coerce — encouraging error on mismatch
- [ ] **`language/variables.md`** — write in full:
  - Implicit declaration on first assignment
  - The deletion rule — with step-by-step trace
  - How to update a variable (delete then re-assign pattern)
  - `del` keyword — creates variable = 0, reset-to-zero behaviour on existing var
  - Scope: `global` is local, `local` is global — table + example
- [ ] **`language/operators.md`** — write in full:
  - Arithmetic table with worked examples (show internal negated values)
  - Comparison inversion table
  - Logical inversion table (`not` is identity, `is` negates)
  - Operator precedence table
  - String `+` removes suffix — with example

#### After Phase 1.7 (Interpreter)

- [ ] **`language/control-flow.md`** — write in full:
  - `if` / `else` with inversion explanation and traced example
  - Block syntax: `}` opens, `{` closes — visual comparison with normal languages
  - `while` — loops while false, exits when true — counted loop example
  - `for` — reverse iteration — traced example
  - `break` continues, `continue` breaks — table
  - `match` / `case` — non-matching cases run — example
- [ ] **`language/functions.md`** — write in full:
  - `call` defines, `define` calls — with annotated example
  - Parameters reversed — with trace showing which arg maps to which param
  - `return` discards, `discard` returns — table + example
  - Recursion support
  - First-class functions — assign `call greet` to variable, call via `define fn(...)`
- [ ] **`language/io.md`** — write in full:
  - `input` prints to stdout — number display, string reversal, bool display
  - `print` reads from stdin — with prompt example
  - `inputln` / `println` variants
  - Interactive session example (user types name, program greets)
- [ ] **`language/error-handling.md`** — write in full:
  - `try` never runs, `except` always runs — with explanation
  - `finally` only runs when skipped
  - `raise` suppresses — with example
  - All error messages are encouraging — show the full table from `SPEC.md §17`
- [ ] **`language/modules.md`** — write in full:
  - `import` removes, `export` loads — with example
  - `wronglib` standard library reference table (all 7 functions, what you expect vs what happens)
- [ ] **`language/reserved-words.md`** — write in full:
  - Full keyword reference table (3 columns: written / expected / actual)
  - Match the table in `SPEC.md §14` exactly, with links to relevant language pages
- [ ] **`examples.md`** — write in full:
  - Hello World — source + annotated trace + output
  - Count from 1 to 5 — source + annotated trace + output
  - FizzBuzz — source + annotated trace + output
  - Function: add two numbers — source + annotated trace + output
  - Reading user input — source + session transcript
- [ ] **`guide/getting-started.md`** — write in full:
  - Prerequisites (Go ≥ 1.22)
  - Install: `go install github.com/KashifKhn/worng/cmd/worng@latest`
  - Verify: `worng version`
  - Write first `.wrg` file (hello world)
  - Run: `worng run hello.wrg`
  - Understand the output (why it's reversed)
  - Next steps: link to language reference
- [ ] **`index.md`** (home page) — replace placeholder with real content:
  - Hero: "Wrong by design. Right by accident."
  - Tagline: one-sentence description
  - Three feature cards: "Everything Inverted", "Comments Are Code", "Bottom to Top"
  - Quick-start code block showing hello world
  - CTA buttons: "Get Started" → guide, "Playground" → playground page
- [ ] **`playground.md`** — update placeholder:
  - Embed `WrongPlayground.vue`
  - Note that WASM runtime is coming in Phase 4; editor and UI are functional

#### After Phase 1.9 (CLI) + 1.10 (Golden Tests)

- [ ] **`guide/getting-started.md`** — add `worng check`, `worng fmt`, `worng --repl` usage
- [ ] Update **`index.md`** quick-start to show the real install command and real CLI output

---

### Phase 2 — LSP Server

**Code milestone:** `worng lsp` operational. Diagnostics, hover, autocomplete, semantic tokens, go-to-definition working.  
**Doc goal:** A dedicated LSP / editor setup guide explaining how to connect any editor to the WORNG LSP.

#### After Phase 2.1–2.8 (full LSP)

- [ ] Create **`guide/lsp.md`** — write in full:
  - What the WORNG LSP provides (diagnostics, hover, autocomplete, semantic tokens, go-to-definition, document symbols)
  - How to start it: `worng lsp` (stdio transport)
  - How to configure a generic LSP client (for editors not covered by Phase 3)
  - JSON-RPC initialize request/response example
  - Troubleshooting: common issues (binary not in PATH, wrong file extension, etc.)
- [ ] Update **`guide/getting-started.md`** — add an "Editor Setup" section linking to `guide/lsp.md` and the editor-specific pages
- [ ] Add **`guide/lsp.md`** to the sidebar under a new "Editor" group

---

### Phase 3 — Editor Integrations

**Code milestone:** VSCode extension published to Marketplace. Neovim plugin installable.  
**Doc goal:** One-page installation guide per editor. No ambiguity — copy-paste instructions that work.

#### After Phase 3.2 (VSCode extension)

- [ ] Create **`guide/vscode.md`** — write in full:
  - Install from Marketplace: search "WORNG" or direct link
  - What you get: syntax highlighting, diagnostics, hover, autocomplete, snippets
  - Configuration options (if any)
  - Screenshot of a `.wrg` file open in VSCode with features active
  - Known issues / limitations

#### After Phase 3.3 (Neovim plugin)

- [ ] Create **`guide/neovim.md`** — write in full:
  - Install with lazy.nvim (primary)
  - Install with packer.nvim (secondary)
  - Manual install (for other plugin managers)
  - Required dependencies: `nvim-lspconfig`, `nvim-treesitter`
  - Default keymaps table
  - Configuration example (`require('worng').setup({})`)
  - Known issues / limitations

#### After Phase 3.1 (tree-sitter grammar)

- [ ] Update **`language/grammar.md`** — add a note linking to the `tree-sitter-worng` repository and explaining how to use it with any tree-sitter-capable editor

#### Sidebar update

- [ ] Add `guide/vscode.md` and `guide/neovim.md` to the sidebar under an "Editor Setup" group:

  ```
  Editor Setup
    LSP Protocol
    VSCode
    Neovim
  ```

---

### Phase 4 — Web Playground

**Code milestone:** WASM binary built. Playground live with real execution.  
**Doc goal:** Playground page fully functional. Update all "coming in Phase 4" stubs to real content.

#### After Phase 4.1 (WASM build) + 4.2 (Playground UI)

- [ ] **`playground.md`** — remove placeholder notice; playground is now fully functional
- [ ] Copy `playground/worng.wasm` to `docs/public/worng.wasm` as part of the build pipeline
- [ ] Update `WrongPlayground.vue`:
  - Wire "Run" button to real `worngRun()` WASM call
  - Wire output panel to display real interpreter output
  - Wire error display to show encouraging WORNG error messages
  - Preset examples dropdown: Hello World, FizzBuzz, Fibonacci, User Input
  - Share button: encode source in `#code=<base64>` URL fragment; load on page open
- [ ] Update **`index.md`** — replace "Playground coming soon" with a live mini-playground embed (hello world example using `input "Hello, World!"`, output `!dlroW ,olleH`; note beneath showing `~` for normal output)
- [ ] Update **`guide/getting-started.md`** — add "Or try it in your browser" section pointing to playground

#### After Phase 4.3 (Deployment)

- [ ] Add playground URL to **`index.md`** hero CTA
- [ ] Verify the share link feature works end-to-end in production

---

### Phase 5 — Polish and Publish

**Code milestone:** `v1.0.0` tagged. GoReleaser producing binaries. Homebrew formula live.  
**Doc goal:** Every page is complete, reviewed, and polished. The site is the public face of WORNG.

#### After Phase 5.1 (Documentation Polish)

- [ ] Proofread every language reference page — check examples compile and produce documented output
- [ ] Add "Edit this page on GitHub" links (VitePress built-in — configure `editLink` in `config.ts`)
- [ ] Add Algolia DocSearch (apply at [docsearch.algolia.com](https://docsearch.algolia.com/apply/) once site is public)
- [ ] Add social meta tags (OG image, Twitter card) — VitePress `head` config
- [ ] Create OG image for the site (1200×630, WORNG logo + tagline)

#### After Phase 5.2 (Release / GoReleaser)

- [ ] Update **`guide/getting-started.md`** — add download options:
  - `go install` (developer install)
  - Pre-built binary from GitHub Releases
  - Homebrew (once formula is live)
- [ ] Add a **`guide/install.md`** page:
  - All install methods in one place
  - Platform-specific notes (Windows PATH setup, etc.)
  - Verify your install section
- [ ] Add `guide/install.md` to sidebar above `getting-started.md`

#### After Phase 5.3 (Homebrew)

- [ ] Update **`guide/install.md`** — add Homebrew install instructions:
  ```
  brew install KashifKhn/worng/worng
  ```

#### After Phase 5.4 (Community)

- [ ] Add a **`community.md`** page:
  - GitHub Discussions link
  - Issue templates overview
  - "Show your WORNG programs" invite
- [ ] Add site footer with: GitHub, Discussions, License (MIT)
- [ ] Update `README.md` — add doc site badge and link

#### Final checklist before v1.0.0 is called done

- [ ] Every page in the sidebar has real content (no stubs remaining)
- [ ] All code examples on every page have been tested against the actual interpreter
- [ ] `npm run build` is zero-warning
- [ ] Lighthouse score ≥ 90 on Performance, Accessibility, Best Practices, SEO
- [ ] Site is indexed by Google (submit sitemap to Search Console)

---

## 9. Page Content Specifications

Detailed content brief for each page. What it must cover, in what order, and what code examples are required.

---

### `index.md` — Home Page

```
Hero section
  H1: "Wrong by Design."
  Subtitle: "WORNG is an esoteric programming language where everything is inverted."
  Subtext: "Only comments execute. Programs run bottom-to-top. + means subtract."
  CTA buttons: [Get Started] [Playground]

Quick example (3-column layout)
  Left:  "You write this" — source code
  Right: "This is what happens" — annotated explanation
  Below: [Try it in the playground →]

Feature cards (3)
  1. Everything Inverted
     "Every operator, keyword, and control flow construct does the opposite of what it says."
  2. Comments Are Code
     "Only commented lines execute. Uncommented lines are ignored. Decoration is not code."
  3. Bottom to Top
     "Programs execute in reverse order. The last line runs first."

Language at a glance (quick reference table — links to full pages)

Footer CTA: "Ready to be wrong?" → [Get Started]
```

---

### `guide/getting-started.md`

```
Prerequisites
Install
  go install (primary method)
  Pre-built binary (link to releases)
  Homebrew (Phase 5)
Verify
  worng version
Write your first program
  Create hello.wrg
  Source with explanation of each line
Run it
  worng run hello.wrg
  Explain the output
Understand what just happened
  Comment principle → execution model link
  Reversal principle → data types link
What's next
  Link to language reference
  Link to examples
  Link to editor setup
```

---

### `language/overview.md`

```
What is WORNG?
  2-3 paragraph overview — not a joke, fully specified, has LSP, editor integrations, playground
The 4 Design Principles
  1. Inversion Principle — every construct does the opposite
  2. Comment Principle — only comments execute
  3. Chaos Principle — when rules conflict, more confusing wins
  4. Encouragement Principle — all errors are positive
A taste of WORNG
  Side-by-side: Python hello world vs WORNG hello world
  Annotated explanation of the differences
Links to every language section
```

---

### `language/execution-model.md`

```
The comment/code rule
  Core rule statement
  The four comment markers — table
  Single-line example
  Block comment example
  Mixed styles example
  No-nesting rule
Execution order
  Bottom-to-top with annotated source example
  Step-by-step trace showing which line runs first
  Diagram: source → preprocessor → reversed executable lines → interpreter
Scope of reversal
  Only top-level statement order is reversed
  Expressions within a statement evaluate left-to-right (normal)
Program termination conditions
```

---

### `language/data-types.md`

```
Numbers
  Integer and float literals
  Internal negation — write 42, store -42
  Worked example chain: write → store → arithmetic → display
  Why this matters for arithmetic results
Strings
  Single and double quotes
  Escape sequences: \n \t \\ \"
  On output: reversed by default — "hello" → "olleh"
  Raw strings: ~"hello" — never reversed, raw flag permanent
  Side-by-side table: regular vs raw, literal vs variable
  String concatenation: + removes suffix (not adds)
Booleans
  true is false, false is true — table
  There is no way to write a literal true
Null
  The one honest literal — not inverted
  Rationale
Type coercion
  None — encouraging error on mismatch
  Show the error message
```

---

### `language/operators.md`

```
Arithmetic operators
  Table: written / actual operation / example / result
  Worked examples showing internal negated values
  Full chain example: 5 + 3 in WORNG prints 2
Comparison operators
  Inversion table
  Example: if x == 5 actually checks x != 5
Logical operators
  not is identity, is negates
  and is or, or is and
  Example combining all three
Operator precedence
  Precedence table (highest to lowest)
  Note: same as conventional — maximises confusion when combined with inverted semantics
String operator
  + removes suffix — with example and "what if not a suffix?" behaviour
```

---

### `language/control-flow.md`

```
Block syntax
  } opens a block, { closes it
  Comparison with conventional languages
  Empty block is valid
  Indentation: required by convention, not enforced
if / else
  Runs when condition is FALSE
  else runs when condition is TRUE
  Traced example (follow a value through the inversion chain)
while
  Loops while condition is FALSE
  Exits when condition becomes TRUE
  Counting loop example with full trace
for
  Reverse iteration
  Example: for x in [1,2,3] prints 3, 2, 1
break and continue
  break continues to next iteration
  continue breaks out of loop
  Table
match / case
  Executes non-matching cases
  Wildcard _ runs when a specific case matches
  Example
```

---

### `language/variables.md`

```
Declaration
  Implicit on first assignment
  No type annotation needed
The deletion rule
  Assigning to an existing variable deletes it
  Step-by-step trace showing creation → deletion → error
How to update a variable
  Pattern: assign once (delete) → assign again (create with new value)
The del keyword
  Creates variable = 0
  On existing variable: deletes then creates with 0 → net reset to zero
Scope
  Variables at top-level are local (not global)
  global keyword makes variable local
  local keyword makes variable global
  Table: keyword / actual scope
  Example showing the inversion
```

---

### `language/functions.md`

```
Definition and calling
  call defines, define calls
  Annotated example side-by-side with conventional equivalent
Parameters
  Reversed on entry
  Traced example: define f(10, 3) → inside f: a=3, b=10
return and discard
  return discards value, returns null
  discard returns value to caller
  Table
  Example showing both
First-class functions
  Assign call funcName to variable
  Call via define varName(...)
Recursion
  Stack overflow: encouraging error message
```

---

### `language/io.md`

```
input — output to stdout
  Prints numbers (negated for display — appears normal)
  Prints strings (reversed by default)
  Prints raw strings (~"...") without reversal
  Prints variables: reversal depends on whether value is raw or regular
  Prints booleans (inverted)
  Examples for each type — regular vs raw side by side
print — read from stdin
  Reads a line, returns value as a regular string
  With prompt string: displays reversed if regular, as-is if raw
  Example: ask name with raw prompt, greet user
  Example: store and re-output user input (reversed by default)
inputln and println
  inputln: print with trailing newline
  println: read, strip trailing newline
  When to use each
```

---

### `language/error-handling.md`

```
try / except
  try never runs (interpreter cannot predict errors)
  except always runs
  Example: what you expect vs what actually happens
finally
  Only runs when skipped by return/break/continue
  Skipped when execution flows into it naturally
  Example showing the inversion
raise
  Suppresses an active exception
  Example
Error messages
  All WORNG errors are encouraging and positive
  Full table from SPEC.md §17
  Show formatted error output example: [W0001] Amazing progress!...
```

---

### `language/modules.md`

```
import and export
  import removes module from namespace
  export loads module
  Example: load → use → remove
wronglib standard library
  Reference table (function / what you expect / what it does)
  Worked example for each function
  Note: math works normally in wronglib — WORNG isn't THAT evil
```

---

### `language/reserved-words.md`

```
Full keyword table
  3 columns: Keyword / Expected behaviour / Actual WORNG behaviour
  Link each keyword to its relevant language page
Identifiers
  Rules for valid identifiers
  Case-sensitive note
Comment markers
  Not keywords — but listed here for completeness
```

---

### `language/grammar.md`

```
EBNF grammar (verbatim from SPEC.md §15)
  Annotated with brief explanations beside each rule
  Link from each production rule to the relevant language section
Token definitions
  NUMBER, STRING, IDENTIFIER, NEWLINE — with regex-style notation
Notes on parsing
  Comment on how the preprocessor feeds into the parser
  Note on non-nesting block delimiters
```

---

### `examples.md`

```
For each example:
  - Title and brief description
  - Source code (syntax-highlighted .wrg block)
  - Annotated trace (what each line does, step by step)
  - Actual output

Examples to include:
  1. Hello World
  2. Count from 1 to 5
  3. FizzBuzz (1–20)
  4. Function: add two numbers
  5. Reading user input
  6. Fibonacci (recursion)
  7. Scope demonstration (global/local inversion)
  8. Error handling (try/except)

[Try in Playground →] link after each example
```

---

### `playground.md`

```
H1: "Playground"
Brief: "Write and run WORNG programs in your browser."

<WrongPlayground /> (full-width Vue component)

Below the playground:
  Tips for first-time WORNG writers
  Link to examples
  Link to language reference
```

---

### `architecture.md`

```
For contributors — summary of ARCHITECTURE.md
  Project structure overview
  Package dependency rules
  Build system (Makefile targets)
  Code generation (Phase 2 LSP proto; diagnostics are hand-maintained)
  VFS abstraction
  Testing strategy (unit, golden, fuzz)
  LSP architecture
  WASM build
[Read full ARCHITECTURE.md on GitHub →] link
```

---

### `roadmap.md`

```
Current status (auto-updated manually per release)
  What version is current, what's in progress
Phase table
  Each phase: goal, status, key deliverable
  Link to GitHub milestone for each phase
Contribution opportunities
  Phases not yet started — link to issues
[View ROADMAP.md on GitHub →] link
```

---

## 10. VitePress Configuration Reference

Skeleton of `docs/.vitepress/config.ts` showing all nav and sidebar entries. To be filled in when the site is scaffolded.

```typescript
import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'WORNG',
  description: 'Wrong by design. Right by accident.',
  base: '/worng/',

  head: [
    ['link', { rel: 'icon', href: '/worng/logo.svg' }],
    ['meta', { property: 'og:title', content: 'WORNG Language' }],
    ['meta', { property: 'og:description', content: 'Wrong by design. Right by accident.' }],
  ],

  themeConfig: {
    logo: '/logo.svg',
    siteTitle: 'WORNG',

    editLink: {
      pattern: 'https://github.com/KashifKhn/worng/edit/main/docs/:path',
      text: 'Edit this page on GitHub',
    },

    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'Language', link: '/language/overview' },
      { text: 'Examples', link: '/examples' },
      { text: 'Playground', link: '/playground' },
      {
        text: 'v1.0.0',
        items: [
          { text: 'Changelog', link: 'https://github.com/KashifKhn/worng/releases' },
          { text: 'Contributing', link: 'https://github.com/KashifKhn/worng/blob/main/CONTRIBUTING.md' },
        ]
      }
    ],

    sidebar: {
      '/guide/': [
        {
          text: 'Guide',
          items: [
            { text: 'Getting Started', link: '/guide/getting-started' },
          ]
        },
        // Phase 2: add { text: 'LSP Protocol', link: '/guide/lsp' }
        // Phase 3: add editor setup group
        // Phase 5: add install page
      ],

      '/language/': [
        {
          text: 'Language Reference',
          items: [
            { text: 'Overview', link: '/language/overview' },
            { text: 'Execution Model', link: '/language/execution-model' },
            { text: 'Data Types', link: '/language/data-types' },
            { text: 'Operators', link: '/language/operators' },
            { text: 'Control Flow', link: '/language/control-flow' },
            { text: 'Variables', link: '/language/variables' },
            { text: 'Functions', link: '/language/functions' },
            { text: 'Input & Output', link: '/language/io' },
            { text: 'Error Handling', link: '/language/error-handling' },
            { text: 'Modules', link: '/language/modules' },
            { text: 'Reserved Words', link: '/language/reserved-words' },
            { text: 'Grammar', link: '/language/grammar' },
          ]
        }
      ],

      '/': [
        {
          text: 'Internals',
          items: [
            { text: 'Architecture', link: '/architecture' },
            { text: 'Roadmap', link: '/roadmap' },
          ]
        }
      ]
    },

    socialLinks: [
      { icon: 'github', link: 'https://github.com/KashifKhn/worng' }
    ],

    footer: {
      message: 'Wrong by design. Right by accident.',
      copyright: 'MIT License — KashifKhn'
    },

    search: {
      provider: 'local'
      // Upgrade to Algolia DocSearch in Phase 5
    }
  }
})
```

---

*WORNG Documentation Website Plan v1.0.0*  
*"If the docs make sense, you've written them wrong."*
