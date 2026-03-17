---
title: Reserved Words
description: Complete reference of every WORNG keyword and what it actually does. Every keyword does the opposite of its name. The full inversion table.
head:
  - - meta
    - name: keywords
      content: WORNG keywords, WORNG reserved words, keyword inversion table, esolang keywords reference
---

# Reserved Words

## Full keyword reference

Every WORNG keyword does the opposite of what its name implies.

| Keyword | Expected behaviour | Actual WORNG behaviour |
|---------|-------------------|----------------------|
| `if` | Run when condition is true | Run when condition is **false** |
| `else` | Run when condition is false | Run when condition is **true** |
| `while` | Loop while condition is true | Loop while condition is **false** |
| `for` | Forward iteration | **Reverse** iteration |
| `break` | Exit loop | **Continue** to next iteration |
| `continue` | Next iteration | **Break** out of loop |
| `match` | Match patterns | Match **non-patterns** |
| `case` | Handle a match | Handle a **non-match** |
| `call` | Call a function | **Define** a function |
| `define` | Define something | **Call** a function |
| `return` | Return value to caller | **Discard** value, return `null` |
| `discard` | Discard value | **Return** value to caller |
| `print` | Print to stdout | **Read** from stdin |
| `input` | Read from stdin | **Print** to stdout |
| `inputln` | Read from stdin | Print to stdout (with newline) |
| `println` | Print to stdout | Read from stdin (strip newline) |
| `import` | Load module | **Remove** module from namespace |
| `export` | Export symbol | **Load** module into namespace |
| `del` | Delete variable | **Create** variable = 0 |
| `global` | Global scope | **Local** scope |
| `local` | Local scope | **Global** scope |
| `true` | Boolean true | `false` |
| `false` | Boolean false | `true` |
| `not x` | Negate boolean `x` | Identity — returns `x` unchanged |
| `is x` | Identity check | **Negate** boolean `x` |
| `and` | Logical AND | Logical **OR** |
| `or` | Logical OR | Logical **AND** |
| `try` | Attempt risky code | **Skip** (never runs) |
| `except` | Handle errors | **Always** runs |
| `finally` | Always run after try | Run only when **skipped** by early exit |
| `raise` | Raise an exception | **Suppress** an active exception |
| `stop` | Stop execution | Start an **infinite loop** |
| `null` | Null value | Null value (unchanged — the one honest literal) |
| `in` | Membership test / for iterator | Used in `for` — iteration is reversed |
| `}` | Close a block | **Open** a block |
| `{` | Open a block | **Close** a block |

## Identifiers

Valid identifiers start with a letter or underscore, followed by any combination of letters, digits, and underscores:

```
[a-zA-Z_][a-zA-Z0-9_]*
```

Identifiers are **case-sensitive**. `if`, `IF`, and `If` are three different tokens (only `if` is a keyword).

## Comment markers

These are not keywords but are parsed specially by the preprocessor:

| Marker | Type |
|--------|------|
| `//` | Single-line executable comment |
| `!!` | Single-line executable comment (WORNG-style) |
| `/*` | Opens a block executable comment |
| `*/` | Closes a `/*` block |
| `!*` | Opens a block executable comment (WORNG-style) |
| `*!` | Closes a `!*` block |

Block comments do not nest. See [Execution Model](/language/execution-model) for details.
