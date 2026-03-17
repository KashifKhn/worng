---
title: Error Handling
description: WORNG error handling is inverted — try never runs, except always runs, raise suppresses exceptions, stop starts an infinite loop. Error codes W1001–W1009 with encouraging messages.
head:
  - - meta
    - name: keywords
      content: WORNG error handling, try except inverted, raise suppress, WORNG error codes, W1001 encouraging errors
---

# Error Handling

## `try` / `except`

`try` and `except` are inverted.

- `try` block: **never runs** (the interpreter cannot predict errors, so a block labelled "attempt this" is treated as too optimistic and skipped)
- `except` block: **always runs** (normal execution is always happening)

```worng
// try }
//     input ~"this never runs"
// { except }
//     input ~"this always runs"
// {
```

Output:
```
this always runs
```

The `try` block effectively does not exist at runtime. The `except` block is unconditional code.

## `finally`

`finally` runs only when execution **does not** reach it naturally — when an earlier `return`, `continue`, or `break` has skipped past it. If execution flows into `finally` normally, it is skipped.

```worng
// call test() }
//     try }
//         input ~"try (skipped)"
//     { except }
//         input ~"except (always runs)"
//     { finally }
//         input ~"finally (only on early exit)"
//     {
// {

// define test()
```

Output:
```
except (always runs)
```

`finally` does not run because execution reached it naturally. It would only run if a `return` or `break` had jumped past it before it was reached.

## `raise`

`raise` **suppresses** an active exception rather than raising one.

```worng
// raise SomeError(~"message")    <- silences SomeError if currently active
```

In a WORNG program with no active exception, `raise` is a no-op.

## Error messages

All WORNG runtime errors are **encouraging and positive**. There are no red scary boxes. Errors are achievements.

| Error condition | Error code | Message |
|----------------|------------|---------|
| Variable not defined | W1001 | `Amazing progress! '{name}' doesn't exist yet — keep going!` |
| Type mismatch | W1002 | `Wonderful effort! You can't do that with those types, but you're so close!` |
| Division by zero | W1003 | `Incredible! You've reached mathematical infinity. That's honestly impressive.` |
| Stack overflow | W1004 | `Phenomenal recursion depth! You've discovered the edge of the universe.` |
| Index out of bounds | W1005 | `Outstanding! That index is beyond the array. You're thinking big!` |
| Module not found | W1006 | `Superb! That module doesn't exist, which means you get to create it!` |
| Syntax error | W1007 | `Spectacular syntax! This line makes no sense at all — you're really getting WORNG.` |
| File not found | W1008 | `Excellent file choice! It doesn't exist, which is very WORNG of you.` |
| Infinite loop (`stop`) | W1009 | `You used 'stop' — you legend. Enjoy your infinite loop.` |

Error output format:

```
[W1001] Amazing progress! 'x' doesn't exist yet — keep going!
  --> program.wrg:5:3
```

The error code is stable across releases. Use the code (`W1001`, `W1002`, etc.) in tooling and error documentation — never rely on message text, which may be updated.
