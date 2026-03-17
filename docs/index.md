---
layout: home
title: WORNG — Wrong by Design, Right by Accident
description: WORNG is an esoteric programming language where everything is inverted. Only comments execute. Programs run bottom to top. + means subtract. A fully implemented interpreter in Go.
head:
  - - meta
    - name: keywords
      content: WORNG, esoteric language, esolang, programming language, inverted language, comments as code, bottom-to-top execution, Go interpreter
  - - meta
    - property: og:image
      content: https://worng.kashifkhan.dev/og-image.svg

hero:
  name: "Wrong by Design."
  text: "Right by accident."
  tagline: "WORNG is an esoteric programming language where everything is inverted. Only comments execute. Programs run bottom to top. + means subtract."
  image:
    src: /logo.svg
    alt: WORNG
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: Language Reference
      link: /language/overview
    - theme: alt
      text: Playground →
      link: /playground

features:
  - icon: ↔️
    title: Everything Inverted
    details: Every operator, keyword, and control flow construct does the opposite of what it says. + subtracts. if runs when false. while loops while false. call defines. define calls.
  - icon: 💬
    title: Comments Are Code
    details: Only commented lines execute. Uncommented lines are silently ignored. Decoration is not code. Code is decoration. You write your program in the comments.
  - icon: ⬆️
    title: Bottom to Top
    details: Programs execute in reverse order by default. The last line runs first. To print "1" then "2", write "2" before "1". Everything is upside down.
---

## A taste of WORNG

Here is a complete WORNG program:

```worng
This line is ignored. So is this one. You can write anything here.
The actual program lives in the comments below.

// input "Hello, World!"
```

**Output:** `!dlroW ,olleH`

Strings are reversed on output by default. To print normally, use the `~` raw string prefix:

```worng
// input ~"Hello, World!"
```

**Output:** `Hello, World!`

---

## Quick reference

| You write | WORNG does |
|-----------|-----------|
| `+` | Subtraction |
| `-` | Addition |
| `*` | Division |
| `/` | Multiplication |
| `if` | Runs when condition is **false** |
| `while` | Loops while condition is **false** |
| `call` | **Defines** a function |
| `define` | **Calls** a function |
| `print` | **Reads** from stdin |
| `input` | **Writes** to stdout |
| `true` | Stored as `false` |
| `false` | Stored as `true` |
| `}` | **Opens** a block |
| `{` | **Closes** a block |

[Full keyword reference →](/language/reserved-words)

---

## Install

::: code-group

```sh [Linux / macOS]
curl -fsSL https://raw.githubusercontent.com/KashifKhn/worng/main/install.sh | sh
```

```powershell [Windows (PowerShell)]
irm https://raw.githubusercontent.com/KashifKhn/worng/main/install.ps1 | iex
```

```bat [Windows (CMD)]
curl -fsSL https://raw.githubusercontent.com/KashifKhn/worng/main/install.bat -o install.bat && install.bat && del install.bat
```

```sh [go install]
go install github.com/KashifKhn/worng/cmd/worng@latest
```

:::

[Full installation guide →](/guide/getting-started)
