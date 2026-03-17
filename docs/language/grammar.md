---
title: Grammar (EBNF)
description: The formal EBNF grammar of the WORNG programming language — the authoritative definition used to implement the recursive-descent parser.
head:
  - - meta
    - name: keywords
      content: WORNG grammar, WORNG EBNF, formal grammar, WORNG parser, esoteric language spec
---

# Grammar (EBNF)

The formal grammar of WORNG. This is the authoritative definition used to implement the parser.

## Productions

```ebnf
program         = line* EOF

line            = ignored_line | exec_line
ignored_line    = (any text not starting with comment_marker) NEWLINE
exec_line       = comment_marker statement NEWLINE
                | block_comment_open statement* block_comment_close

comment_marker      = "//" | "!!"
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
                | RAW_STRING
                | "true"
                | "false"
                | "null"
                | IDENTIFIER
                | "(" expression ")"
                | array_literal
                | func_call_expr

array_literal   = "[" (expression ("," expression)*)? "]"
func_call_expr  = "define" IDENTIFIER "(" arg_list? ")"
```

## Token definitions

```
NUMBER      = [0-9]+ ("." [0-9]+)?
STRING      = '"' [^"]* '"' | "'" [^']* "'"
RAW_STRING  = "~" STRING
IDENTIFIER  = [a-zA-Z_] [a-zA-Z0-9_]*
NEWLINE     = "\n" | "\r\n"
```

## Notes

**Preprocessor vs parser:** The preprocessor runs before the lexer. It filters source lines, keeping only those beginning with `//`, `!!`, or inside `/* */` / `!* *!` blocks. The parser never sees uncommented lines.

**Block delimiters:** `}` opens a block, `{` closes it. The grammar rule `block = exec_line*` means a block is a sequence of executable lines between the inverted brace pair.

**No nesting of block comments:** The grammar treats `block_comment_open` and `block_comment_close` as non-nesting. The first `*/` or `*!` always ends the current block comment.

**`func_call_expr`:** `define` appears in both `func_call_stmt` (statement context) and `func_call_expr` (expression context, when a function call is used as a value). Both use the same production.
