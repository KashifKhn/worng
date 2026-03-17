---
title: Playground
description: Write and run WORNG programs live in your browser. Try the interactive playground — no install required.
layout: page
head:
  - - meta
    - name: keywords
      content: WORNG playground, run WORNG online, WORNG browser, interactive esolang, try WORNG
---

# Playground

Write and run WORNG programs in your browser. Use `Ctrl+Enter` / `Cmd+Enter` to run.

<WrongPlayground />

---

## Tips for first-time WORNG writers

- Every line must start with `//` or `!!` to execute. Any other line is silently ignored.
- Programs execute **bottom to top** by default. Write your setup code below its first use.
- `+` subtracts. `-` adds. `*` divides. `/` multiplies.
- `}` opens a block. `{` closes it.
- `call` defines a function. `define` calls it.
- `input` prints to stdout. `print` reads from stdin.
- Regular strings are **reversed on output**. Use `~"..."` for a raw string that prints as written.

## Learn more

- [Language Reference →](/language/overview)
- [Annotated Examples →](/examples)
- [Getting Started →](/guide/getting-started)
