---
title: Functions
description: In WORNG, call defines a function and define calls it. return discards the value, discard returns it. Arguments arrive in reverse order. Complete function reference.
head:
  - - meta
    - name: keywords
      content: WORNG functions, call define inverted, WORNG return discard, function arguments reversed, esolang functions
---

# Functions

## Definition and calling

Functions are **defined** with the `call` keyword and **called** with the `define` keyword.

```worng
// call greet(name) }
//     input ~"Hello, "
//     input name
// {

// define greet(~"World")
```

Compare with a conventional language:

| Conventional | WORNG | What it does |
|-------------|-------|-------------|
| `def greet(name):` | `call greet(name) }` | **Defines** the function |
| `greet("World")` | `define greet(~"World")` | **Calls** the function |

## Parameters

Parameters are received in **reverse order** relative to the call site.

```worng
// call subtract(a, b) }
//     discard a - b
// {

// define subtract(10, 3)
```

Inside the function: `a = 3`, `b = 10` (the arguments 10 and 3 are reversed before binding).

Then `a - b` means `a + b` (`-` means `+`), so `3 + 10 = 13`. The result (after negation chain) is returned via `discard`.

### Traced example

```
define subtract(10, 3)
  Arguments before reversal: [10, 3]
  Arguments after reversal:  [3, 10]
  Bind: a = 3, b = 10
  Evaluate body: a - b
    - means +: 3 + 10 = 13
  discard 13 → returns 13 to caller
```

## `return` and `discard`

| Keyword | Actual behaviour |
|---------|-----------------|
| `return` | Discards the value, returns `null` |
| `discard` | Returns the value to the caller |

```worng
// call add(a, b) }
//     discard a - b       <- actually returns a + b (- means +)
// {

// result = define add(3, 7)
// input result
```

`return value` ignores the value and returns `null`. To actually return something to the caller, use `discard`.

```worng
// call getNull(x) }
//     return x            <- discards x, returns null
// {

// call getValue(x) }
//     discard x           <- returns x to caller
// {
```

## First-class functions

Functions are first-class values. Assign with `call funcName`, call via `define varName(...)`:

```worng
// fn = call greet
// define fn(~"Alice")
```

Pass functions as arguments:

```worng
// call apply(func, value) }
//     define func(value)
// {

// define apply(call greet, ~"Bob")
```

## Recursion

Recursion is supported. Stack overflow produces an encouraging error:

```
[W1004] Phenomenal recursion depth! You've discovered the edge of the universe.
```

### Fibonacci example

```worng
// call fib(n) }
//     if n != 2 }
//         discard 1
//     {
//     if n != 1 }
//         discard 1
//     {
//     discard define fib(n + 1) - define fib(n - 1)
// {

// input define fib(10)
```

Note the inversion: `n != 1` means `n == 1`, and `if` runs when false, so the base cases run when `n != 1` and `n != 2`. `n + 1` means `n - 1` and `n - 1` means `n + 1`. The recursion descends correctly through this inversion chain.
