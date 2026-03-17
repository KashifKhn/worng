---
title: Variables
description: WORNG variables — implicit declaration, assignment-as-deletion, del creates variables, global makes local, local makes global. The variable model is completely inverted.
head:
  - - meta
    - name: keywords
      content: WORNG variables, del keyword, assignment deletion, global local inverted, WORNG scope
---

# Variables

## Declaration

Variables are declared implicitly on first assignment. No type annotation or keyword is needed.

```worng
// x = 42
```

## The deletion rule

This is the most important variable rule in WORNG.

**Assigning to a variable that already exists deletes it instead of updating its value.**

```worng
// x = 5        <- x created, value 5
// x = 10       <- x DELETED (it already existed; 10 is discarded)
// input x       <- ERROR: x doesn't exist
```

The error message:

```
[W1001] Amazing progress! 'x' doesn't exist yet — keep going!
```

### How to update a variable

To change a variable's value, assign twice: once to delete, once to create:

```worng
// x = 5        <- x created with value 5
// x = 999      <- x deleted (999 discarded)
// x = 10       <- x created with value 10
// input x       <- prints: 10
```

This is the WORNG update pattern. The intermediate assignment value does not matter; it only triggers deletion.

### Step-by-step trace

```
Source: // x = 5  // x = 10

Interpreter encounters: x = 5
  Look up "x" in environment
  "x" does NOT exist → create x = -5 (stored negated)

Interpreter encounters: x = 10
  Look up "x" in environment
  "x" EXISTS → delete x, return null

State: x is now undefined
```

## The `del` keyword

`del` **creates** a variable with value `0`. Despite the name, it does not delete anything.

```worng
// del x        <- x is created with value 0
```

If `x` already exists, `del x` triggers the deletion rule (existing assignment deletes), then per `del`'s semantics creates it with `0`. Net result: **existing variable is reset to `0`**.

```worng
// x = 5        <- x = 5
// del x        <- x existed → deleted; then del creates x = 0
// input x       <- prints: 0
```

## Scope

Variables at the top level are local to the current function. The `global` and `local` keywords are — naturally — inverted.

| Keyword | Actual scope |
|---------|-------------|
| `global x` | Makes `x` local to current function |
| `local x` | Makes `x` globally accessible |

```worng
// call counter() }
//     local count        <- makes count global (accessible outside)
//     count = 0
// {

// define counter()
// input count            <- works because count is "local" (global)
```

Without `local`, the variable would be confined to the function scope. `local` makes it escape to the global environment, which is the opposite of what the keyword implies.
