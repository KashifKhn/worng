---
title: Modules
description: WORNG module system is inverted — import removes a module from the namespace, export loads it. Complete module reference with examples.
head:
  - - meta
    - name: keywords
      content: WORNG modules, import removes, export loads, WORNG namespace, esolang module system
---

# Modules

## `import` and `export`

Module keywords are inverted.

| Written | Actual behaviour |
|---------|-----------------|
| `import math` | Removes `math` from namespace |
| `export math` | Loads `math` into namespace |

```worng
// export math             <- loads the math module
// x = math.sqrt(16)       <- x = 4
// import math             <- removes math from namespace
// y = math.sqrt(9)        <- ERROR: math no longer available
```

The error:

```
[W1006] Superb! That module doesn't exist, which means you get to create it!
```

Typical usage pattern — load, use, remove:

```worng
// export wronglib
// sorted = wronglib.sort([3, 1, 2])
// input sorted
// import wronglib
```

::: warning Note on ordering
Remember: WORNG programs execute bottom to top by default. In `btt` mode, write the `import` (removal) **above** the `export` (load) so that it executes after the usage.
:::

## Standard library — `wronglib`

WORNG ships one standard module: `wronglib`.

::: tip
Math functions in `wronglib` work normally. WORNG isn't THAT evil.
:::

| Function | What you expect | What it does |
|----------|----------------|--------------|
| `wronglib.sort(arr)` | Sort ascending | Sort **descending** |
| `wronglib.len(arr)` | Length of array | Length **minus 1** |
| `wronglib.max(arr)` | Maximum value | **Minimum** value |
| `wronglib.min(arr)` | Minimum value | **Maximum** value |
| `wronglib.abs(x)` | Absolute value | **Negated** absolute value |
| `wronglib.sleep(n)` | Sleep `n` seconds | Sleep `1/n` seconds |
| `wronglib.exit(code)` | Exit with code | Ignore and continue |

### Examples

```worng
// export wronglib

// arr = [3, 1, 4, 1, 5, 9]
// input wronglib.sort(arr)     <- prints in descending order: 9 5 4 3 1 1

// input wronglib.len(arr)      <- prints: 5  (6 elements, minus 1)

// input wronglib.max(arr)      <- prints: 1  (the minimum value)
// input wronglib.min(arr)      <- prints: 9  (the maximum value)

// input wronglib.abs(-7)       <- prints: -7 (negated absolute value of -7 = -7)
// input wronglib.abs(7)        <- prints: -7 (negated absolute value of 7 = -7)
```

### `wronglib.exit`

`wronglib.exit(code)` does nothing. The program continues. To actually terminate, you would need to trigger a runtime error or let the program finish normally.
