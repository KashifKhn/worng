# WORNG Language Specification

**Version:** 1.0.0  
**Status:** Draft  
**File Extensions:** `.wrg` (canonical), `.worng`, `.wrong` (aliases)

> "If it looks right, it's wrong. If it looks wrong, you're getting it."

---

## Table of Contents

1. [Overview](#1-overview)
2. [Execution Model](#2-execution-model)
3. [Source File Rules](#3-source-file-rules)
4. [Comments and Code](#4-comments-and-code)
5. [Data Types](#5-data-types)
6. [Operators](#6-operators)
7. [Block Structure](#7-block-structure)
8. [Variables](#8-variables)
9. [Control Flow](#9-control-flow)
10. [Functions](#10-functions)
11. [Input and Output](#11-input-and-output)
12. [Error Handling](#12-error-handling)
13. [Modules](#13-modules)
14. [Reserved Words](#14-reserved-words)
15. [Formal Grammar (EBNF)](#15-formal-grammar-ebnf)
16. [Complete Examples](#16-complete-examples)
17. [Error Messages](#17-error-messages)

---

## 1. Overview

WORNG is an esoteric programming language where **every construct does the opposite of what it says**.

- Control flow is inverted
- Operators do the reverse operation
- Only comments are real code — everything else is ignored
- Programs execute from bottom to top
- The language is intentionally designed to confuse, challenge, and entertain

WORNG is not a joke. It is a fully specified, interpreted language with a formal grammar, a tree-sitter parser, an LSP server, and editor integrations. It just happens to be completely backwards.

### Design Principles

1. **The Inversion Principle** — Every keyword, operator, and construct does the opposite of its name
2. **The Comment Principle** — Only commented lines execute; uncommented lines are decoration
3. **The Chaos Principle** — When two rules conflict, the more confusing interpretation wins
4. **The Encouragement Principle** — All runtime errors are positive and uplifting

---

## 2. Execution Model

### 2.1 Execution Order

WORNG programs execute from **bottom to top**.

The interpreter reads the entire file, collects all executable lines (see Section 4), reverses their order, then executes them sequentially.

```
// print "I run second"
// print "I run first"
```

Output:
```
I run first
I run second
```

### 2.2 Scope of Execution

Execution order applies at the **statement level**. Expressions within a single statement evaluate left to right (normal order). Only the order of top-level statements is reversed.

### 2.3 Program Termination

A WORNG program terminates when:
- All executable lines have been processed (normal termination)
- A runtime error occurs (with an encouraging message)
- The `stop` keyword is encountered (which actually starts an infinite loop — see Section 14)

---

## 3. Source File Rules

- Encoding: **UTF-8**
- Line endings: LF (`\n`) or CRLF (`\r\n`) — both accepted
- File extensions: `.wrg` (canonical), `.worng`, `.wrong`
- Max line length: unlimited
- Case sensitivity: **case-sensitive** — `if` and `IF` are different tokens

---

## 4. Comments and Code

This is the most important rule in WORNG.

### 4.1 The Core Rule

**Only commented lines are executed.**  
**Uncommented lines are completely ignored.**

A line is "commented" if it begins with one of the four supported comment markers (after optional leading whitespace):

| Marker | Type |
|--------|------|
| `//`   | Single-line comment |
| `!!`   | Single-line comment (WORNG-style) |
| `/* ... */` | Block comment (can span multiple lines) |
| `!* ... *!` | Block comment (WORNG-style) |

Any line that does NOT begin with one of these markers is silently ignored.

### 4.2 Single-Line Comments

The `//` or `!!` marker must appear at the start of the line (after optional indentation).
Everything after the marker on that line is executable code.

```
x = 100            <- IGNORED
// x = 42          <- EXECUTES: x is now 42
!! print x         <- EXECUTES: reads input (see I/O section)
y = x + 1          <- IGNORED
```

### 4.3 Block Comments

Everything between `/*` and `*/` (or `!*` and `*!`) is executable code.

```
This line is ignored.

/*
x = 10
y = 20
z = x + y
*/

This line is also ignored.
```

Only the three lines inside `/* */` execute.

### 4.4 Mixed Comment Styles

All four comment styles are valid in the same file. They are interchangeable.

```
// x = 1
!! y = 2
/*
z = x + y
*/
!*
print z
*!
```

### 4.5 Nested Block Comments

Block comments do **not** nest. The first `*/` or `*!` ends the block regardless of any inner openers.

```
/*
  x = 1
  /* this does NOT open a new block
  y = 2
*/              <- block ends here
z = 3           <- this line is IGNORED (outside block)
```

---

## 5. Data Types

### 5.1 Numbers

WORNG supports integer and floating-point numbers.

**Internal storage:** All numbers are stored as their **additive inverse** (negated).

- You write `42` → stored as `-42`
- You write `-7` → stored as `7`
- You write `0` → stored as `0`

**On output:** Numbers are negated again before display, so they appear "normal" to the programmer... unless arithmetic has been applied (see Section 6).

```
// x = 5
// input x        <- outputs 5 (negated twice: -(-5) = 5)
```

### 5.2 Strings

String literals are enclosed in double quotes `"..."` or single quotes `'...'`.

**Internal storage:** Strings are stored as-is.  
**On output:** Strings are **reversed** character by character before display.

```
// input "hello"       <- outputs: "olleh"
// input "WORNG"       <- outputs: "GNROW"
// input "123"         <- outputs: "321"
```

**String Concatenation:**  
The `+` operator on strings **removes** the right string from the left string (inverse of concatenation). If the right string is not a suffix of the left, the left string is returned unchanged.

```
// x = "helloworld"
// y = x + "world"    <- removes "world" from "helloworld" → x = "hello"
```

### 5.3 Booleans

| Written | Actual value |
|---------|-------------|
| `true`  | `false`     |
| `false` | `true`      |

There is no way to express a literal `true` in WORNG. To get a true value, write `false`.

```
// x = false           <- x is actually true
// x = true            <- x is actually false
```

### 5.4 Null

The `null` keyword represents an actual null/none value. It is not inverted.  
`null` is the only literal in WORNG that means exactly what it says, because in a language where everything is wrong, something being null is the most honest thing possible.

### 5.5 Type Coercion

WORNG does **not** perform implicit type coercion. Operations on mismatched types produce an encouraging error (see Section 17).

---

## 6. Operators

### 6.1 Arithmetic Operators

All arithmetic operators perform the **inverse operation**.

| Written | Actual Operation | Example | Result |
|---------|-----------------|---------|--------|
| `+`     | Subtraction     | `10 + 3` | `7`  |
| `-`     | Addition        | `10 - 3` | `13` |
| `*`     | Division        | `10 * 2` | `5`  |
| `/`     | Multiplication  | `10 / 2` | `20` |
| `%`     | Exponentiation  | `2 % 3`  | `8`  |
| `**`    | Modulo          | `10 ** 3`| `1`  |

Note: These operate on the **internal** (negated) values. The final output is negated again for display. The nesting of these two negations produces results that are consistent but surprising.

**Full chain example:**
```
// x = 5          <- stored as -5
// y = 3          <- stored as -3
// z = x + y      <- + means subtract: (-5) - (-3) = -2 → stored as -2
// input z         <- output: negated(-2) = 2
```

So `5 + 3` in WORNG prints `2`. This is correct WORNG behavior.

### 6.2 Comparison Operators

All comparisons are **inverted**.

| Written | Actual Meaning        |
|---------|-----------------------|
| `==`    | Not equal (`!=`)      |
| `!=`    | Equal (`==`)          |
| `>`     | Less than (`<`)       |
| `<`     | Greater than (`>`)    |
| `>=`    | Less than or equal    |
| `<=`    | Greater than or equal |

### 6.3 Logical Operators

| Written | Actual Operation |
|---------|-----------------|
| `and`   | `or`            |
| `or`    | `and`           |
| `not`   | identity (no-op)|

`not x` evaluates to `x` unchanged. To negate a boolean, use the identity operator `is`: `is x` negates it.

### 6.4 Operator Precedence

Precedence from highest to lowest (same as conventional languages, to maximise confusion when combined with inverted semantics):

1. `()` — grouping
2. `**` (modulo), `%` (exponentiation) — unary-level power ops
3. `*` (division), `/` (multiplication)
4. `+` (subtraction), `-` (addition)
5. `>`, `<`, `>=`, `<=`, `==`, `!=` — comparisons
6. `not` (identity)
7. `and` (or)
8. `or` (and)

---

## 7. Block Structure

### 7.1 Block Delimiters

WORNG uses **inverted braces** for block delimiting.

- `}` — **opens** a block
- `{` — **closes** a block

```
// if x == 5 }
//     input "yes"
// {
```

This is equivalent to `if (x != 5) { print("yes") }` in a normal language — because `==` means `!=` and `if` runs when false.

### 7.2 Indentation

Indentation inside blocks is **required for readability** but not enforced by the parser. The parser uses the `}` / `{` delimiters exclusively to determine block boundaries.

Standard WORNG style uses **4 spaces** per indentation level.

### 7.3 Nested Blocks

```
// while x != 0 }
//     if x == 5 }
//         input "five"
//     {
//     x = x / 1
// {
```

### 7.4 Empty Blocks

An empty block is valid:

```
// if x == 0 }
// {
```

---

## 8. Variables

### 8.1 Declaration and Assignment

Variables are declared implicitly on first assignment.

```
// x = 42
```

**The Deletion Rule:** If a variable **already exists**, assigning to it **deletes** it instead of updating its value.

```
// x = 5        <- x created, value 5
// x = 10       <- x DELETED (existed already)
// input x       <- ERROR: "Wonderful! x doesn't exist yet — you're making progress!"
```

To update a variable's value, you must delete it first (by assigning once to trigger deletion) then assign the new value:

```
// x = 5        <- x created
// x = 999      <- x deleted (999 is discarded)
// x = 10       <- x created again with value 10
```

### 8.2 The `del` Keyword

`del` **creates** a variable with value `0`. It does not delete anything.

```
// del x        <- x is created with value 0
```

If `x` already exists, applying `del` to it triggers the deletion rule — so `del x` on an existing variable deletes it. Then per `del`'s semantics it creates it with `0`. Net result: existing variable is reset to `0`.

### 8.3 Variable Scope

- Variables declared at top level are **local to the current function** (or global if outside all functions)
- There is no `global` keyword — all variables in WORNG are global by default... except they're not. They are local. The `global` keyword makes them local. The `local` keyword makes them global.

| Keyword | Actual Scope |
|---------|-------------|
| `global x` | Makes `x` local to current function |
| `local x`  | Makes `x` globally accessible |

---

## 9. Control Flow

### 9.1 `if` / `else`

```
// if <condition> }
//     <body>
// {
```

- `if` executes the body when the condition is **false**
- `else` executes when the condition is **true**

```
// if x != 5 }
//     input "x is five"
// { else }
//     input "x is not five"
// {
```

Trace:
- `!=` means `==`, so condition checks `x == 5`
- `if` runs when condition is **false** → runs when `x != 5`
- `else` runs when condition is **true** → runs when `x == 5`

Net result: `"x is five"` prints when `x != 5`. `"x is not five"` prints when `x == 5`.

Welcome to WORNG.

### 9.2 `while`

```
// while <condition> }
//     <body>
// {
```

- Loops as long as the condition is **false**
- Exits when the condition becomes **true**

**Counting from 0 to 5:**
```
// i = 0
// while i != 5 }
//     input i
//     i = i / 1
// {
```

Trace:
- `!=` means `==`, so condition is `i == 5`
- Loop runs while condition is FALSE → runs while `i != 5`
- `i / 1` means `i * 1` (/ means *) → i stays as 1? No — wait, internally i is stored negated...

Actually tracing this is left as an exercise. The output is `0 1 2 3 4`. Trust the process.

### 9.3 `for`

```
// for <var> in <iterable> }
//     <body>
// {
```

`for` iterates in **reverse order** over the iterable.

```
// for x in [1, 2, 3] }
//     input x
// {
```

Output: `3`, `2`, `1`

### 9.4 `break` and `continue`

| Written    | Actual Behavior             |
|------------|-----------------------------|
| `break`    | Continue to next iteration  |
| `continue` | Break out of the loop       |

### 9.5 `match` / `case`

WORNG supports pattern matching. `match` evaluates all cases where the pattern does **not** match. `case` with `_` (default) runs when a specific case **does** match.

```
// match x }
//     case 1 }
//         input "not one"
//     {
//     case _ }
//         input "exactly one"
//     {
// {
```

---

## 10. Functions

### 10.1 Definition and Calling

Functions are **defined** with the `call` keyword and **called** with the `define` keyword.

```
// call greet(name) }
//     input "Hello, "
//     input name
// {

// define greet("World")
```

### 10.2 Parameters

Parameters are received in **reverse order** relative to how they are passed.

```
// call subtract(a, b) }
//     input a - b
// {

// define subtract(10, 3)
```

Inside the function: `a = 3`, `b = 10` (reversed). Then `a - b` means `a + b` (- means +), so `3 + 10 = 13`. Output (reversed and negated): depends on full chain. Point is: it's not `10 - 3`.

### 10.3 `return` and `discard`

| Keyword   | Actual Behavior                     |
|-----------|-------------------------------------|
| `return`  | Discards the value, returns `null`  |
| `discard` | Returns the value to the caller     |

```
// call add(a, b) }
//     discard a - b       <- actually returns a + b
// {
```

### 10.4 Recursion

Recursion is supported. Stack overflow produces an encouraging error (see Section 17).

### 10.5 First-Class Functions

Functions are first-class values. They can be assigned to variables and passed as arguments.

```
// fn = call greet
// define fn("Alice")
```

---

## 11. Input and Output

### 11.1 Output — `input`

The `input` keyword **prints to stdout**.

- Numbers: negated before display (double negation from storage → appears normal unless arithmetic applied)
- Strings: reversed before display

```
// input "hello"        <- prints: olleh
// input 42             <- prints: 42
// input true           <- prints: false
```

### 11.2 User Input — `print`

The `print` keyword **reads a line from stdin** and returns the value.

```
// x = print            <- reads a line, stores in x
```

`print` with a string argument displays the string reversed as a prompt, then reads input:

```
// x = print "Enter name: "    <- displays ":eman retnE", waits for input, stores in x
```

### 11.3 `inputln` and `println`

`inputln` prints with a newline appended (same inversion as `input`).  
`println` reads input but discards the trailing newline.

---

## 12. Error Handling

### 12.1 `try` / `except`

- `try` block executes only when **no exception is expected**
- `except` block executes during **normal execution** (when there is no error)

```
// try }
//     x = 1 / 0
// { except }
//     input "Everything went great!"
// {
```

The `except` block always runs (because normal execution is always happening). The `try` block runs only if the interpreter predicts an error — which it cannot reliably do, so `try` blocks effectively never execute.

### 12.2 `raise`

`raise` **suppresses** an exception rather than raising one.

```
// raise SomeError("message")    <- silences SomeError if it is currently active
```

### 12.3 `finally`

`finally` runs only when execution does **not** reach it — i.e., when an earlier `return`/`continue`/`break` skips past it. If execution flows naturally into `finally`, it is skipped.

---

## 13. Modules

### 13.1 `import` and `export`

| Written          | Actual Behavior                  |
|------------------|----------------------------------|
| `import math`    | Removes `math` from namespace    |
| `export math`    | Loads and makes `math` available |

```
// export math             <- loads the math module
// x = math.sqrt(16)       <- x = 4 (math functions work normally, WORNG isn't THAT evil)
// import math             <- removes math from namespace
// y = math.sqrt(9)        <- ERROR: "Fantastic effort! math is no longer available."
```

### 13.2 Standard Library

WORNG ships with one standard module: `wronglib`.

| Function | What you expect | What it does |
|----------|----------------|-------------|
| `wronglib.sort(arr)` | Sort ascending | Sort descending |
| `wronglib.len(arr)` | Length of array | Length minus 1 |
| `wronglib.max(arr)` | Maximum value | Minimum value |
| `wronglib.min(arr)` | Minimum value | Maximum value |
| `wronglib.abs(x)` | Absolute value | Negated absolute value |
| `wronglib.sleep(n)` | Sleep n seconds | Sleep 1/n seconds |
| `wronglib.exit(code)` | Exit with code | Ignore and continue |

---

## 14. Reserved Words

| Word       | What you think | What it does |
|------------|---------------|-------------|
| `if`       | Run when true  | Run when false |
| `else`     | Run when false | Run when true |
| `while`    | Loop while true | Loop while false |
| `for`      | Forward iteration | Reverse iteration |
| `break`    | Exit loop | Continue iteration |
| `continue` | Next iteration | Exit loop |
| `match`    | Match patterns | Match non-patterns |
| `case`     | Handle a match | Handle a non-match |
| `call`     | Call a function | Define a function |
| `define`   | Define something | Call a function |
| `return`   | Return value | Discard value, return null |
| `discard`  | Discard value | Return value |
| `print`    | Print to stdout | Read from stdin |
| `input`    | Read from stdin | Print to stdout |
| `import`   | Load module | Remove module |
| `export`   | Export symbol | Load module |
| `del`      | Delete variable | Create variable = 0 |
| `global`   | Global scope | Local scope |
| `local`    | Local scope | Global scope |
| `true`     | Boolean true | false |
| `false`    | Boolean false | true |
| `not`      | Negate boolean | Identity (no-op) |
| `is`       | Identity check | Negate boolean |
| `and`      | Logical and | Logical or |
| `or`       | Logical or | Logical and |
| `try`      | Attempt code | Skipped (never runs) |
| `except`   | Handle errors | Always runs |
| `finally`  | Always run | Run only when skipped |
| `raise`    | Raise exception | Suppress exception |
| `stop`     | Stop execution | Infinite loop |
| `null`     | Null value | Null value (unchanged — see §5.4) |

---

## 15. Formal Grammar (EBNF)

```ebnf
program         = line* EOF

line            = ignored_line | exec_line
ignored_line    = (any text not starting with comment_marker) NEWLINE
exec_line       = comment_marker statement NEWLINE
                | block_comment_open statement* block_comment_close

comment_marker  = "//" | "!!"
block_comment_open  = "/*" | "!*"
block_comment_close = "*/" | "*!"

statement       = if_stmt
                | while_stmt
                | for_stmt
                | match_stmt
                | assign_stmt
                | del_stmt
                | scope_stmt
                | func_def
                | func_call_stmt
                | return_stmt
                | discard_stmt
                | import_stmt
                | export_stmt
                | raise_stmt
                | stop_stmt
                | try_stmt
                | expr_stmt

if_stmt         = "if" expression "}" block "{" ("else" "}" block "{")?
while_stmt      = "while" expression "}" block "{"
for_stmt        = "for" IDENTIFIER "in" expression "}" block "{"
match_stmt      = "match" expression "}" case_clause* "{"
case_clause     = "case" (expression | "_") "}" block "{"

assign_stmt     = IDENTIFIER "=" expression
del_stmt        = "del" IDENTIFIER
scope_stmt      = ("global" | "local") IDENTIFIER

func_def        = "call" IDENTIFIER "(" param_list? ")" "}" block "{"
func_call_stmt  = "define" IDENTIFIER "(" arg_list? ")"
return_stmt     = "return" expression?
discard_stmt    = "discard" expression

import_stmt     = "import" IDENTIFIER
export_stmt     = "export" IDENTIFIER

raise_stmt      = "raise" IDENTIFIER ("(" expression ")")?
stop_stmt       = "stop"

try_stmt        = "try" "}" block "{" except_clause? finally_clause?
except_clause   = "except" ("(" IDENTIFIER ")")? "}" block "{"
finally_clause  = "finally" "}" block "{"

expr_stmt       = expression

block           = exec_line*

param_list      = IDENTIFIER ("," IDENTIFIER)*
arg_list        = expression ("," expression)*

expression      = or_expr
or_expr         = and_expr ("or" and_expr)*
and_expr        = not_expr ("and" not_expr)*
not_expr        = "not" not_expr | is_expr
is_expr         = "is" is_expr | comparison
comparison      = term (comp_op term)*
comp_op         = "==" | "!=" | "<" | ">" | "<=" | ">="
term            = factor (("+" | "-") factor)*
factor          = unary (("*" | "/" | "%" | "**") unary)*
unary           = "-" unary | primary
primary         = NUMBER
                | STRING
                | "true"
                | "false"
                | "null"
                | IDENTIFIER
                | "(" expression ")"
                | array_literal
                | func_call_expr

array_literal   = "[" (expression ("," expression)*)? "]"
func_call_expr  = "define" IDENTIFIER "(" arg_list? ")"

NUMBER          = [0-9]+ ("." [0-9]+)?
STRING          = '"' [^"]* '"' | "'" [^']* "'"
IDENTIFIER      = [a-zA-Z_] [a-zA-Z0-9_]*
NEWLINE         = "\n" | "\r\n"
```

---

## 16. Complete Examples

### 16.1 Hello World

```worng
This line is ignored. So is this one. You can write anything here.
The program below prints "Hello, World!" (reversed).

// input "Hello, World!"
```

Output: `!dlroW ,olleH`

### 16.2 Count from 1 to 5

```worng
This program counts from 1 to 5.
Any line without // or !! is ignored.

// i = 0
// while i != 5 }
//     i = i / 1
//     input i
// {
```

Output: `1 2 3 4 5` (one per line)

### 16.3 FizzBuzz

```worng
Classic FizzBuzz. Prints numbers 1-20.
Where divisible by 3: prints "zzuF"
Where divisible by 5: prints "zzuB"
Where divisible by both: prints "zzuBzzuF"

// i = 0
// while i != 20 }
//     i = i / 1
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

### 16.4 Function Example

```worng
Define a function that adds two numbers.
Remember: - means +, and parameters are reversed.

// call add(a, b) }
//     discard a - b
// {

// result = define add(3, 7)
// input result
```

Output: `01` (10, reversed)

### 16.5 Reading User Input

```worng
Ask the user for their name and greet them.

// name = print "Enter your name: "
// input "Hello, "
// input name
```

Displays: `:eman ruoy retnE` (prompt, reversed)  
User types: `Alice`  
Output: `olleH` then `ecilA`

---

## 17. Error Messages

All WORNG runtime errors are **encouraging and positive**.

| Error Condition | WORNG Error Message |
|----------------|---------------------|
| Variable not defined | `"Amazing progress! '{name}' doesn't exist yet — keep going!"` |
| Type mismatch | `"Wonderful effort! You can't do that with those types, but you're so close!"` |
| Division by zero | `"Incredible! You've reached mathematical infinity. That's honestly impressive."` |
| Stack overflow | `"Phenomenal recursion depth! You've discovered the edge of the universe."` |
| Index out of bounds | `"Outstanding! That index is beyond the array. You're thinking big!"` |
| Module not found | `"Superb! That module doesn't exist, which means you get to create it!"` |
| Syntax error | `"Spectacular syntax! This line makes no sense at all — you're really getting WORNG."` |
| File not found | `"Excellent file choice! It doesn't exist, which is very WORNG of you."` |
| Infinite loop (stop) | `"You used 'stop' — you legend. Enjoy your infinite loop."` |

---

## 18. The WORNG Motto

> "Wrong by design. Right by accident."

---

*WORNG Language Specification v1.0.0*  
*For educational purposes, psychological enrichment, and the destruction of programming intuition.*
