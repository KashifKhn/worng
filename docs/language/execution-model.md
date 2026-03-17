---
title: Execution Model
description: How WORNG programs run — only comments execute, lines run bottom to top by default, and what btt vs ttb execution order means.
head:
  - - meta
    - name: keywords
      content: WORNG execution model, bottom to top, btt ttb, comments as code, WORNG interpreter
---

# Execution Model

## The comment/code rule

This is the single most important rule in WORNG.

**Only commented lines execute. Uncommented lines are completely ignored.**

A line is "commented" if it begins with one of four markers (after optional leading whitespace):

| Marker | Type |
|--------|------|
| `//` | Single-line comment |
| `!!` | Single-line comment (WORNG-style) |
| `/* ... */` | Block comment |
| `!* ... *!` | Block comment (WORNG-style) |

Any line that does not begin with one of these markers is silently discarded before the program runs.

```worng
x = 100            <- IGNORED
// x = 42          <- EXECUTES: x is 42
!! print x         <- EXECUTES: reads from stdin
y = x + 1          <- IGNORED
```

Only the two commented lines run. The rest never exist as far as the interpreter is concerned.

### Block comments

Everything between `/*` and `*/` (or `!*` and `*!`) is executable code:

```worng
This line is ignored.

/*
x = 10
y = 20
z = x + y
*/

This line is also ignored.
```

Only the three lines inside `/* */` execute.

### Mixed styles

All four comment styles are valid in the same file and are interchangeable:

```worng
// x = 1
!! y = 2
/*
z = x + y
*/
!*
input z
*!
```

### No nesting

Block comments do **not** nest. The first `*/` or `*!` ends the block:

```worng
/*
  x = 1
  /* this does NOT open a new block
  y = 2
*/           <- block ends here
z = 3        <- this line is IGNORED (outside the block)
```

## Execution order

### Bottom-to-top (default)

WORNG reads the entire source file, collects all executable lines in **source order**, parses them into an AST, then executes top-level statements in **reverse order**.

```worng
// input ~"I run second"
// input ~"I run first"
```

Output:
```
I run first
I run second
```

The bottom statement executes first. This is `btt` mode and is the default.

### Top-to-bottom (optional)

Pass `--order=ttb` to execute statements in source order:

```bash
worng run --order=ttb program.wrg
```

```worng
// input ~"I run first"
// input ~"I run second"
```

Output (with `--order=ttb`):
```
I run first
I run second
```

### Scope of reversal

Execution order reversal applies **only to top-level statements**. Expressions within a single statement still evaluate left-to-right. Inside a block (`if`, `while`, etc.), statements execute in the order they appear within the block.

### Authoring by mode

| Goal | Write in `btt` (default) | Write in `ttb` |
|------|--------------------------|----------------|
| Initialize before use | Put initializer **below** first use | Put initializer **above** first use |
| Define function before call | Put `call` **below** `define` | Put `call` **above** `define` |
| Print something after a block | Put trailing `input` **above** the block | Put trailing `input` **below** the block |

Example — print `1` then `done` in both modes:

**`btt` source (default):**
```worng
// input ~"done"
// input 1
```

**`ttb` source:**
```worng
// input 1
// input ~"done"
```

Both produce:
```
1
done
```

## Pipeline

```
Source file (.wrg)
      │
      ▼
PREPROCESSOR
  Reads the file
  Keeps only commented lines
  Strips comment markers
  Preserves source order
      │ []string (executable lines)
      ▼
LEXER
  Tokenizes each line
  Produces token stream with position info
      │ []Token
      ▼
PARSER
  Recursive descent
  Produces Abstract Syntax Tree (AST)
      │ *ProgramNode
      ▼
INTERPRETER
  Schedules top-level statements (btt or ttb)
  Walks AST, applies inversion rules
  Produces output
      │
      ▼
   stdout
```

## Program termination

A WORNG program terminates when:

- All executable lines have been processed (normal termination)
- A runtime error occurs (with an encouraging message — see [Error Messages](/language/error-handling))
- The `stop` keyword is encountered — which actually **starts** an infinite loop
