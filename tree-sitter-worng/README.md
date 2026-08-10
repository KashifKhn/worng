# tree-sitter-worng

Tree-sitter grammar for the WORNG programming language.

The Go lexer and parser remain authoritative for execution. This grammar is
for incremental editor parsing, highlighting, indentation, and folding.

## Development

Install the Tree-sitter CLI, then run:

```bash
tree-sitter generate
tree-sitter test
```

The generated parser under `src/` is committed so editor integrations do not
need to generate code during installation.

WORNG's inverted delimiters are represented intentionally:

- `}` is captured as an opening block delimiter
- `{` is captured as a closing block delimiter
- `//` and `!!` lines are parsed as executable code
- `/* ... */` and `!* ... *!` blocks are parsed as executable code
