---
title: Getting Started
description: Install WORNG on Linux, macOS, or Windows and write your first inverted program. Covers the CLI, REPL, and execution modes.
head:
  - - meta
    - name: keywords
      content: WORNG install, WORNG getting started, WORNG CLI, esoteric language setup, Go esolang
---

# Getting Started

## Install

::: code-group

```bash [Linux / macOS]
curl -fsSL https://raw.githubusercontent.com/KashifKhn/worng/main/install.sh | sh
```

```powershell [Windows (PowerShell)]
irm https://raw.githubusercontent.com/KashifKhn/worng/main/install.ps1 | iex
```

```bat [Windows (CMD)]
curl -fsSL https://raw.githubusercontent.com/KashifKhn/worng/main/install.bat -o install.bat && install.bat && del install.bat
```

```bash [go install]
go install github.com/KashifKhn/worng/cmd/worng@latest
```

:::

The installer downloads a pre-built binary from [GitHub Releases](https://github.com/KashifKhn/worng/releases) and adds it to your PATH automatically.

### Optional flags

| Flag | Effect |
|------|--------|
| `--version 0.1.0` | Install a specific version |
| `--no-modify-path` | Skip PATH modification |

**Linux / macOS with flags:**

```bash
curl -fsSL https://raw.githubusercontent.com/KashifKhn/worng/main/install.sh | sh -s -- --version 0.1.0
```

**PowerShell with flags:**

```powershell
& ([scriptblock]::Create((irm https://raw.githubusercontent.com/KashifKhn/worng/main/install.ps1))) -Version 0.1.0
```

### Verify

```bash
worng version
```

Expected output:

```
worng v0.1.0
```

## Write your first program

Create a file named `hello.wrg`:

```worng
This line is ignored. Uncommented lines are decoration.
The program lives below.

// input ~"Hello, World!"
```

Two things are happening here:

1. The first two lines have no comment marker — they are **silently ignored**. Only lines starting with `//` or `!!` execute.
2. `input` prints to stdout (not reads from it). `~"Hello, World!"` is a raw string — it prints as-is without reversal.

## Run it

```bash
worng run hello.wrg
```

Output:

```
Hello, World!
```

## Without the `~` prefix

Remove the `~` and the string reverses:

```worng
// input "Hello, World!"
```

Output:

```
!dlroW ,olleH
```

Strings are **reversed on output by default**. The `~` prefix marks a raw string that is never reversed. This is WORNG.

## Execution order

By default, WORNG programs execute **bottom to top**. Write this:

```worng
// input ~"I run second"
// input ~"I run first"
```

Output:

```
I run first
I run second
```

The bottom statement runs first. To switch to top-to-bottom order, use `--order=ttb`:

```bash
worng run --order=ttb hello.wrg
```

## CLI commands

```
worng run [--order=btt|ttb] <file>    Run a .wrg file
worng run [--order=btt|ttb] --repl    Interactive REPL
worng check [--order=btt|ttb] <file>  Parse without running
worng fmt <file>                      Format in-place
worng version                         Print version
```

`btt` (bottom-to-top) is the default execution order. `ttb` (top-to-bottom) runs statements in source order.

## Interactive REPL

Start the REPL for quick experimentation:

```bash
worng run --repl
```

```
WORNG v0.1.0 — Type // or !! followed by WORNG code.
>>> // x = 5
>>> // input x
5
>>> this line is ignored
>>> // input x - 3
8
```

Non-comment lines are silently ignored in the REPL, consistent with file execution.

## What's next

- [Language Reference](/language/overview) — understand every feature
- [Execution Model](/language/execution-model) — understand bottom-to-top execution
- [Data Types](/language/data-types) — numbers, strings, booleans, null
- [Examples](/examples) — annotated programs
