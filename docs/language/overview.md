---
title: Language Overview
description: A high-level introduction to the WORNG programming language — what it is, why it exists, and the core inversion principle that governs every construct.
head:
  - - meta
    - name: keywords
      content: WORNG overview, esoteric programming language, inversion principle, esolang design, wrong by design
---

# Overview

## What is WORNG?

WORNG is an esoteric programming language where **every construct does the opposite of what it says**. Control flow is inverted. Operators perform the reverse operation. Only comments are real code — everything else is ignored. Programs execute from bottom to top by default.

WORNG is not a joke. It is a fully specified, interpreted language with a formal grammar, a recursive-descent parser, an LSP server, and editor integrations for VSCode and Neovim. It just happens to be completely backwards.

The language is designed to challenge programming intuition, teach you to read code carefully, and produce a specific kind of confusion that is, ultimately, educational. Every rule is consistent. Nothing is random. Once you understand the inversion pattern, WORNG programs become predictable — predictably wrong, but predictable.

## The four design principles

### 1. The Inversion Principle

Every keyword, operator, and construct does the opposite of its name.

`if` runs when the condition is false. `while` loops while false. `for` iterates in reverse. `call` defines a function. `define` calls it. `print` reads input. `input` prints output. `true` is `false`. `false` is `true`.

The inversion is total and consistent. There are no exceptions (except `null`, because in a language where everything is wrong, something being null is the most honest thing possible).

### 2. The Comment Principle

Only commented lines execute. Uncommented lines are silently ignored.

```worng
This line is ignored.
So is this one.
x = 100                 <- also ignored
// x = 42               <- executes: x is now 42
```

A "commented line" is any line beginning with `//`, `!!`, or content inside `/* ... */` or `!* ... *!` block markers. Everything else is decoration.

This inverts the normal relationship between code and comments. In WORNG, your comments are your program.

### 3. The Chaos Principle

When two rules conflict, the more confusing interpretation wins.

This principle rarely activates in practice — the inversion rules are carefully designed to be consistent — but it exists as a tiebreaker. WORNG chooses confusion.

### 4. The Encouragement Principle

All runtime errors are positive and uplifting.

```
[W1001] Amazing progress! 'x' doesn't exist yet — keep going!
[W1003] Incredible! You've reached mathematical infinity. That's honestly impressive.
[W1004] Phenomenal recursion depth! You've discovered the edge of the universe.
```

Errors in WORNG are not failures. They are achievements.

## WORNG is not Python

A side-by-side comparison:

**Python hello world:**
```python
print("Hello, World!")
```

**WORNG hello world:**
```worng
This line is ignored. Your program lives in the comments.

// input ~"Hello, World!"
```

- `input` prints (not reads)
- `~"..."` is a raw string that does not reverse on output
- Non-comment lines are not executed
- The program has exactly one executable line

Without the `~`, strings reverse on output:

```worng
// input "Hello, World!"
```

Output: `!dlroW ,olleH`

## Language sections

| Section | What it covers |
|---------|----------------|
| [Execution Model](/language/execution-model) | Comment/code rule, bottom-to-top order |
| [Data Types](/language/data-types) | Numbers, strings, booleans, null |
| [Operators](/language/operators) | Arithmetic, comparison, logical |
| [Control Flow](/language/control-flow) | if/else, while, for, break/continue, match |
| [Variables](/language/variables) | Assignment, deletion rule, scope |
| [Functions](/language/functions) | call/define, reversed params, return/discard |
| [Input & Output](/language/io) | input, print, raw strings |
| [Error Handling](/language/error-handling) | try/except/finally/raise |
| [Modules](/language/modules) | import/export, wronglib stdlib |
| [Reserved Words](/language/reserved-words) | Full keyword reference table |
| [Grammar](/language/grammar) | Formal EBNF grammar |
