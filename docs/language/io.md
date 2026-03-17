---
title: Input & Output
description: In WORNG, input prints to stdout and print reads from stdin. Strings are reversed on output unless prefixed with ~. Complete I/O reference with examples.
head:
  - - meta
    - name: keywords
      content: WORNG input output, print reads stdin, input writes stdout, raw string tilde, WORNG io
---

# Input & Output

## `input` — output to stdout

The `input` keyword **prints to stdout**. Its name implies the opposite.

### Numbers

Numbers are negated before display (double negation from storage — appears normal unless arithmetic has been applied):

```worng
// input 42             <- prints: 42
// input -7             <- prints: -7
```

### Strings

Regular strings are **reversed** on output:

```worng
// input "hello"        <- prints: olleh
// input "WORNG"        <- prints: GNROW
```

Raw strings (prefixed with `~`) are printed as-is:

```worng
// input ~"hello"       <- prints: hello
// input ~"WORNG"       <- prints: WORNG
```

### Variables

The raw flag travels with the value. If a variable holds a raw string, it prints without reversal:

```worng
// x = "hello"
// input x              <- prints: olleh (regular string)

// y = ~"hello"
// input y              <- prints: hello (raw string)
```

### Booleans

`input` prints the inverted value:

```worng
// input true           <- prints: false
// input false          <- prints: true
```

### Side-by-side reference

| Code | Output |
|------|--------|
| `input "hello"` | `olleh` |
| `input ~"hello"` | `hello` |
| `input 42` | `42` |
| `input true` | `false` |
| `input false` | `true` |
| `x = "hi"; input x` | `ih` |
| `x = ~"hi"; input x` | `hi` |

## `print` — read from stdin

The `print` keyword **reads a line from stdin**. Its name implies the opposite.

```worng
// x = print            <- reads a line, stores result in x
```

### With a prompt

`print` accepts an optional string argument as a prompt. Regular strings are reversed before display; raw strings are printed as-is:

```worng
// x = print "Enter name: "    <- displays ":eman retnE", waits for input
// x = print ~"Enter name: "   <- displays "Enter name: ", waits for input
```

### Interactive example

Ask the user for their name and greet them:

```worng
// name = print ~"Enter your name: "
// input ~"Hello, "
// input name
```

Session:
```
Enter your name: Alice
Hello,
ecilA
```

The name is stored as a regular string (read from stdin), so it reverses on output. To print it normally, wrap it in `~`:

```worng
// name = print ~"Enter your name: "
// raw_name = ~name
// input ~"Hello, "
// input raw_name
```

Session:
```
Enter your name: Alice
Hello,
Alice
```

## `inputln` and `println`

`inputln` is `input` with a newline appended. Use it when you want output followed by a line break.

`println` reads input and strips the trailing newline from the result.

| Keyword | Behaviour |
|---------|-----------|
| `input` | Print to stdout |
| `inputln` | Print to stdout, append newline |
| `print` | Read from stdin |
| `println` | Read from stdin, strip trailing newline |
