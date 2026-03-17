---
title: Data Types
description: WORNG's five data types — numbers stored negated, strings reversed on output, booleans inverted, null, and arrays. Every type is wrong in the right way.
head:
  - - meta
    - name: keywords
      content: WORNG data types, inverted numbers, reversed strings, raw strings, WORNG booleans, esolang types
---

# Data Types

WORNG has five data types: numbers, strings, booleans, null, and arrays. Four of them are inverted.

## Numbers

WORNG supports integer and floating-point number literals.

**Internal storage:** All numbers are stored as their **additive inverse** (negated).

| You write | Stored as |
|-----------|-----------|
| `42` | `-42` |
| `-7` | `7` |
| `3.14` | `-3.14` |
| `0` | `0` |

**On output:** Numbers are negated again before display, so they appear normal unless arithmetic has been applied.

```worng
// x = 5
// input x        <- outputs: 5  (stored -5, displayed -(-5) = 5)
```

The double negation makes single values display as expected. The surprise comes when you do arithmetic:

```worng
// x = 5          <- stored as -5
// y = 3          <- stored as -3
// z = x + y      <- + means subtract: (-5) - (-3) = -2
// input z         <- output: -(-2) = 2
```

`5 + 3` in WORNG prints `2`. This is correct WORNG behaviour.

See [Operators](/language/operators) for the full arithmetic inversion table.

## Strings

String literals use double quotes `"..."` or single quotes `'...'`.

**Escape sequences:**

| Sequence | Meaning |
|----------|---------|
| `\n` | newline |
| `\t` | tab |
| `\\` | backslash |
| `\"` | double quote |
| `\'` | single quote |

**Storage:** Strings are stored as-is after escape processing.

**Output:** Strings are **reversed character by character** before display.

```worng
// input "hello"       <- outputs: olleh
// input "WORNG"       <- outputs: GNROW
// input "123"         <- outputs: 321
```

### Raw strings — the `~` prefix

Prefix a string literal with `~` to mark it as a **raw string**. Raw strings are never reversed on output.

```worng
// input ~"hello"      <- outputs: hello
// input ~"WORNG"      <- outputs: WORNG
```

The `~` prefix is permanent. It travels with the value through variable assignment:

```worng
// x = ~"hello"        <- x is a raw string
// input x             <- outputs: hello  (NOT reversed)
```

A regular string assigned to a variable always reverses on output:

```worng
// x = "hello"         <- x is a reversing string
// input x             <- outputs: olleh
```

### Side-by-side comparison

| Code | Output |
|------|--------|
| `input "hello"` | `olleh` |
| `input ~"hello"` | `hello` |
| `x = "hello"; input x` | `olleh` |
| `x = ~"hello"; input x` | `hello` |

### String `+` (not concatenation)

The `+` operator on strings **removes** the right string from the left string as a suffix (inverse of concatenation). If the right string is not a suffix of the left, the left string is returned unchanged.

```worng
// x = "helloworld"
// y = x + "world"     <- removes "world" suffix -> y = "hello"
// input y             <- outputs: olleh
```

The raw flag of the result follows the left operand:

```worng
// x = ~"helloworld"
// y = x + ~"world"    <- y = ~"hello" (raw flag from left operand)
// input y             <- outputs: hello
```

## Booleans

| You write | Actual value |
|-----------|-------------|
| `true` | `false` |
| `false` | `true` |

There is no way to write a literal `true` in WORNG. To get a true value, write `false`.

```worng
// x = false           <- x is actually true
// x = true            <- x is actually false
```

`input` with a boolean prints the inverted value:

```worng
// input true          <- prints: false
// input false         <- prints: true
```

## Null

`null` represents an absent value. It is the one literal in WORNG that is not inverted.

```worng
// x = null
// input x             <- prints: null
```

`null` means exactly what it says, because in a language where everything is wrong, something being null is the most honest thing possible.

## Type coercion

WORNG does **not** perform implicit type coercion. Operations on mismatched types produce an encouraging error:

```
[W1002] Wonderful effort! You can't do that with those types, but you're so close!
```
