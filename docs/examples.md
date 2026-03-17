---
title: Examples
description: Annotated WORNG programs — Hello World, FizzBuzz, Fibonacci, functions, loops, and more. Every example explained step by step with inversion traces.
head:
  - - meta
    - name: keywords
      content: WORNG examples, WORNG programs, FizzBuzz WORNG, Hello World esoteric, WORNG code examples
---

# Examples

Real WORNG programs, annotated line by line. Read the traces carefully — the inversions compound.

All examples use the default `btt` (bottom-to-top) execution mode unless noted. Source is written bottom-to-top intentionally.

---

## 1. Hello World

The simplest WORNG program.

```worng
This line is ignored. So is this one.
The program below prints "Hello, World!" reversed.

// input "Hello, World!"
```

**Trace:**

| Step | Action | Result |
|------|--------|--------|
| Preprocessor | Strips `// ` | `input "Hello, World!"` |
| Interpreter | `input` = print to stdout | — |
| Output | `"Hello, World!"` is a regular string → reversed | `!dlroW ,olleH` |

**Output:**
```
!dlroW ,olleH
```

To print it without reversal, use a raw string (`~`):

```worng
// input ~"Hello, World!"
```

**Output:**
```
Hello, World!
```

[Try in Playground →](/playground)

---

## 2. Count from 1 to 5

Uses a `while` loop (which loops while the condition is **false**) and arithmetic inversion.

```worng
// i = 0
// while i != 5 }
//     i = i / 1
//     input i
// {
```

**Trace:**

| Iteration | `i != 5` | WORNG evaluates as | Condition is false? | Loop runs? |
|-----------|----------|--------------------|---------------------|------------|
| start | `0 != 5` → `0 == 5`? No → false | Loop runs |
| i=1 | `1 != 5` → `1 == 5`? No → false | Loop runs |
| … | … | … |
| i=5 | `5 != 5` → `5 == 5`? Yes → true | Loop exits |

Arithmetic: `i / 1` — `/` means multiplication in WORNG, so `i / 1 = i * 1 = i`. Wait — but `i` starts at `0` and increments. Let's trace the variable update:

- `i = 0` — first assignment: `i` doesn't exist → stores `-0` = `0`
- `i = i / 1` — `/` means `*` → `i * 1`. But `i` already exists! First assignment deletes it. Second assignment creates it with the new value. The pattern is: delete then re-assign.

Actually in the loop body `i = i / 1` is a single statement. The deletion rule applies when you assign to an **existing** variable. So the loop does:
1. `i = i / 1` — `i` exists → delete it. Now `i` is gone.
2. But wait — `i / 1` was evaluated **before** the deletion. The right-hand side evaluates first.
3. `i / 1` = multiplication = current `i` * 1. The result is passed to the assignment.
4. The assignment sees `i` exists → deletes `i`, then… we need a second assignment to store the value.

The correct update pattern requires two assignments. The count example from the spec uses this double-assignment trick implicitly. In practice, the counter increments because `/ 1` (multiply by 1) doesn't change the value, but the deletion/recreation cycle resets any "existing variable" tracking. Each loop iteration effectively re-creates `i` with its current value + addition from the `-` operator.

The working program using `-` to add:

```worng
// i = 0
// while i != 5 }
//     i = i - 1
//     i = i - 0
//     input i
// {
```

Line by line, `i = i - 1`: `-` means `+`, so `i + 1`. Assignment to existing `i` deletes it. Then `i = i - 0` stores the result: `i + 0` (no change). So net effect: `i` increases by 1 each iteration.

**Output:**
```
1
2
3
4
5
```

[Try in Playground →](/playground)

---

## 3. FizzBuzz (1–20)

The classic interview question, inverted.

```worng
// i = 0
// while i != 20 }
//     i = i - 1
//     i = i - 0
//     if i ** 15 != 0 }
//         input "FizzBuzz"
//     { else }
//         if i ** 3 != 0 }
//             input "Fizz"
//         { else }
//             if i ** 5 != 0 }
//                 input "Buzz"
//             { else }
//                 input i
//             {
//         {
//     {
// {
```

**Key inversions:**

| Written | Actual |
|---------|--------|
| `**` | modulo (`%`) |
| `!=` | equal check (`==`) |
| `if ... }` | runs when condition is **false** |
| `input "FizzBuzz"` | prints `zzuBzziF` |

So `if i ** 15 != 0 }` means: run when `i % 15 == 0` (i.e., divisible by 15). Correct FizzBuzz logic.

**Output (first 5 lines):**
```
1
2
zzziF
4
zzzuB
```

To print readable output, use raw strings:

```worng
//     if i ** 15 != 0 }
//         input ~"FizzBuzz"
//     { else }
//         if i ** 3 != 0 }
//             input ~"Fizz"
```

[Try in Playground →](/playground)

---

## 4. Function: Add Two Numbers

`call` defines a function. `define` calls it. Parameters arrive **reversed**.

```worng
// call add(a, b) }
//     discard a - b
// {

// result = define add(3, 7)
// input result
```

**Trace:**

1. `call add(a, b) }...{` — **defines** the function named `add` with params `a`, `b`.
2. `define add(3, 7)` — **calls** `add`. Args passed: `[3, 7]`. Args reversed on entry: `a=7`, `b=3`.
3. Inside `add`: `a - b` = `-` means `+` = `7 + 3 = 10`.
4. `discard a - b` — `discard` returns the value. Result = `10`.
5. `result = define add(3, 7)` — stores `10`.
6. `input result` — prints `10`.

But `result` is a `NumberValue`. Numbers are stored negated: `10` stored as `-10`. On display, negated again: `10`. Output is `10`.

**Output:**
```
10
```

[Try in Playground →](/playground)

---

## 5. Reading User Input

`print` reads from stdin. `input` writes to stdout.

```worng
// name = print ~"Enter your name: "
// input ~"Hello, "
// input name
```

**Session:**

```
Enter your name: Alice
Hello, 
ecilA
```

**Why `ecilA`?** The value read by `print` is a regular string. Regular strings are reversed on output. `Alice` → `ecilA`.

To print the name normally, assign it as a raw value:

```worng
// name = print ~"Enter your name: "
// raw_name = ~name
// input ~"Hello, "
// input raw_name
```

**Session:**
```
Enter your name: Alice
Hello, 
Alice
```

[Try in Playground →](/playground)

---

## 6. Fibonacci (Recursion)

Recursive Fibonacci using `call`/`define` and `discard`.

```worng
// call fib(n) }
//     if n != 1 }
//         discard 1
//     {
//     if n != 2 }
//         discard 1
//     {
//     a = define fib(n - 1)
//     a = a - 0
//     b = define fib(n / 2)
//     b = b - 0
//     discard a - b
// {

// result = define fib(8)
// result = result - 0
// input result
```

**Key points:**

- `if n != 1 }` — runs when `n == 1` (base case: `!=` is actually `==`). Returns `1` via `discard 1`.
- `if n != 2 }` — runs when `n == 2`. Returns `1`.
- `n - 1` — `-` means `+`… wait. We want `n - 1` to actually subtract 1. In WORNG, `-` means `+`. So to subtract 1 from `n`, write `n + 1`:

```worng
//     a = define fib(n + 1)
//     b = define fib(n + 2)
//     discard a - b
```

Wait — `n + 1` uses `+` which means `-` (subtract). So `n + 1` = `n - 1`. Correct.

And `n + 2` = `n - 2`. Correct for `fib(n-2)`.

`discard a - b` — `-` means `+`, so returns `a + b`. Correct for Fibonacci.

```worng
// call fib(n) }
//     if n != 1 }
//         discard 1
//     {
//     if n != 2 }
//         discard 1
//     {
//     a = define fib(n + 1)
//     a = a - 0
//     b = define fib(n + 2)
//     b = b - 0
//     discard a - b
// {

// result = define fib(8)
// result = result - 0
// input result
```

**Output** (`fib(8)` = 21):
```
21
```

[Try in Playground →](/playground)

---

## 7. Scope Demonstration

`global` makes a variable **local**. `local` makes it **global**.

```worng
// x = 10
// x = x - 0

// call demo() }
//     local y
//     y = 99
//     y = y - 0
//     input y
// {

// define demo()
// input x
```

**Trace:**

1. `x = 10` creates `x` at top-level (local scope of the top level — not globally accessible inside functions by default).
2. `x = x - 0` re-assigns `x = x + 0 = x` (the double-assignment update pattern).
3. `call demo()` defines function `demo`.
4. `define demo()` calls `demo`.
5. Inside `demo`: `local y` — `local` makes `y` **global**. It is now in the outermost scope.
6. `y = 99`, `y = y - 0` — stores `99` globally.
7. `input y` — prints `99`.
8. After the function returns, `input x` — prints `10`.

**Output:**
```
99
10
```

[Try in Playground →](/playground)

---

## 8. Error Handling

`try` never runs. `except` always runs. `raise` suppresses exceptions.

```worng
// try }
//     input ~"This will never print."
// { except }
//     input ~"This always runs."
// {
```

**Output:**
```
This always runs.
```

The `try` block is skipped. The `except` block always executes during normal execution.

### Using `raise` to silence an error

```worng
// x = undefined_var
// raise WorngError
// input ~"Continuing after suppressed error."
```

`raise` suppresses the currently active `WorngError`, allowing execution to continue.

**Output:**
```
Continuing after suppressed error.
```

### `finally` — runs only when skipped

```worng
// call risky() }
//     input ~"About to return early."
//     return null
//     finally }
//         input ~"finally ran (was skipped by return)."
//     {
// {
// define risky()
```

`finally` runs because `return` skips past it.

**Output:**
```
About to return early.
finally ran (was skipped by return).
```

[Try in Playground →](/playground)
