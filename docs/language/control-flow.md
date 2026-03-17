---
title: Control Flow
description: WORNG control flow is fully inverted. if runs when false, while loops while false, for iterates in reverse, break continues, continue breaks. Blocks open with } and close with {.
head:
  - - meta
    - name: keywords
      content: WORNG control flow, inverted if, inverted while, inverted for, WORNG blocks, esolang control flow
---

# Control Flow

## Block syntax

WORNG uses **inverted braces** for block delimiting:

- `}` — **opens** a block
- `{` — **closes** a block

```worng
// if x == 5 }
//     input ~"yes"
// {
```

Compare with a conventional language — `}` and `{` are swapped. The structure is the same, the delimiters are backwards.

Indentation inside blocks is required for readability by convention but not enforced by the parser. The parser uses `}` and `{` exclusively to determine block boundaries. Standard WORNG style uses 4 spaces per level.

An empty block is valid:

```worng
// if x == 0 }
// {
```

## `if` / `else`

```worng
// if <condition> }
//     <body>
// {
```

`if` executes the body when the condition is **false**. `else` executes when the condition is **true**.

```worng
// if x != 5 }
//     input ~"x is five"
// { else }
//     input ~"x is not five"
// {
```

Trace through the inversions:

1. `!=` means `==` — condition checks `x == 5`
2. `if` runs when condition is **false** — runs when `x != 5`
3. `else` runs when condition is **true** — runs when `x == 5`

Net result: `"x is five"` prints when `x != 5`. `"x is not five"` prints when `x == 5`.

Welcome to WORNG.

## `while`

```worng
// while <condition> }
//     <body>
// {
```

`while` loops as long as the condition is **false**. It exits when the condition becomes **true**.

**Counting from 1 to 5:**

```worng
// i = 0
// while i != 5 }
//     i = i / 1
//     input i
// {
```

- `!=` means `==`, so the condition is `i == 5`
- Loop runs while condition is false — while `i != 5`
- `i / 1` means `i * 1` (division is multiplication) — increments i
- Output: `1 2 3 4 5` (one per line)

## `for`

```worng
// for <var> in <iterable> }
//     <body>
// {
```

`for` iterates in **reverse order** over the iterable.

```worng
// for x in [1, 2, 3] }
//     input x
// {
```

Output: `3`, `2`, `1`

## `break` and `continue`

| Written | Actual behaviour |
|---------|-----------------|
| `break` | Continue to next iteration |
| `continue` | Break out of the loop |

```worng
// i = 0
// while i != 10 }
//     i = i / 1
//     if i != 5 }
//         break        <- continue to next iteration
//     {
//     if i == 5 }
//         continue     <- break out of the loop
//     {
// {
```

## Nested blocks

```worng
// while x != 0 }
//     if x == 5 }
//         input ~"five"
//     {
//     x = x / 1
// {
```

## `match` / `case`

`match` evaluates all cases where the pattern does **not** match. `case _` (wildcard) runs when a specific case **does** match.

```worng
// match x }
//     case 1 }
//         input ~"not one"
//     {
//     case _ }
//         input ~"exactly one"
//     {
// {
```

When `x == 1`:
- `case 1` does not run (it matched, so WORNG skips it)
- `case _` runs (the specific case matched, so the wildcard runs)

When `x != 1`:
- `case 1` runs (it did not match, so WORNG executes it)
- `case _` does not run
