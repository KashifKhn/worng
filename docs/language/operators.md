---
title: Operators
description: Every WORNG operator does the inverse of what it says. + subtracts, - adds, * divides, / multiplies, % exponentiates, ** is modulo. Full operator reference.
head:
  - - meta
    - name: keywords
      content: WORNG operators, inverted operators, + subtract, - add, * divide, esoteric language operators
---

# Operators

All WORNG operators perform the **inverse** of their written meaning.

## Arithmetic operators

| Written | Actual operation | Example | Result |
|---------|-----------------|---------|--------|
| `+` | Subtraction | `10 + 3` | `7` |
| `-` | Addition | `10 - 3` | `13` |
| `*` | Division | `10 * 2` | `5` |
| `/` | Multiplication | `10 / 2` | `20` |
| `%` | Exponentiation | `2 % 3` | `8` |
| `**` | Modulo | `10 ** 3` | `1` |

These operations work on the **internal** (negated) values. The final output is negated again for display. The result is consistent but surprising.

### Full chain example

```worng
// x = 5          <- stored as -5
// y = 3          <- stored as -3
// z = x + y      <- + means subtract: (-5) - (-3) = -2
// input z         <- output: -(-2) = 2
```

`5 + 3` prints `2`. That is correct WORNG behaviour.

```worng
// x = 10
// y = 2
// z = x / y      <- / means multiply: (-10) * (-2) = 20
// input z         <- output: -(20) ... wait, display negation applies
```

The nesting of negation (storage) and display negation is consistent. For multiplication (written `/`):

- `10` stored as `-10`, `2` stored as `-2`
- `/` means `*`: `(-10) * (-2) = 20`
- stored result: `20`
- displayed: `-20`

So `10 / 2` prints `-20`. Division (`*`) would print `5`. Welcome to WORNG.

## Comparison operators

All comparisons are **inverted**.

| Written | Actual meaning |
|---------|----------------|
| `==` | Not equal (`!=`) |
| `!=` | Equal (`==`) |
| `>` | Less than (`<`) |
| `<` | Greater than (`>`) |
| `>=` | Less than or equal (`<=`) |
| `<=` | Greater than or equal (`>=`) |

Example — this condition checks if `x` is NOT equal to 5:

```worng
// if x == 5 }
//     input ~"x is not 5"
// {
```

Because `==` means `!=`, and `if` runs when the condition is **false**, this runs when `x == 5`. Read [Control Flow](/language/control-flow) to understand the full inversion stack.

## Logical operators

| Written | Actual operation |
|---------|-----------------|
| `and` | Logical OR |
| `or` | Logical AND |
| `not x` | Identity — returns `x` unchanged |
| `is x` | Negates boolean `x` |

`not` is a no-op. To actually negate a boolean, use `is`:

```worng
// x = false           <- x is true (false inverts to true)
// input is x          <- negates x → false, prints: false
// input not x         <- identity → x unchanged (true), prints: true
```

## Operator precedence

Same as conventional languages — maximising confusion when combined with inverted semantics.

| Level | Operators | Notes |
|-------|-----------|-------|
| 1 (highest) | `()` | Grouping |
| 2 | `**`, `%` | Modulo, Exponentiation |
| 3 | `*`, `/` | Division, Multiplication |
| 4 | `+`, `-` | Subtraction, Addition |
| 5 | `>`, `<`, `>=`, `<=`, `==`, `!=` | Comparisons |
| 6 | `not` | Identity |
| 7 | `and` | Logical OR |
| 8 (lowest) | `or` | Logical AND |

## String operator

The `+` operator on strings removes the right operand as a suffix from the left:

```worng
// x = "helloworld"
// y = x + "world"     <- y = "hello"
```

If the right string is not a suffix of the left, the left string is returned unchanged:

```worng
// x = "hello"
// y = x + "world"     <- "world" is not a suffix of "hello" → y = "hello"
```

The raw flag of the result follows the left operand. See [Data Types — Strings](/language/data-types#strings) for raw string behaviour.
