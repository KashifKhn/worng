# WORNG

> "Wrong by design. Right by accident."

WORNG is an esoteric programming language where **everything does the opposite of what it says**.

- Only **comments** are real code — everything else is ignored
- Programs execute **bottom to top**
- `+` subtracts, `-` adds, `*` divides, `/` multiplies
- `if` runs when the condition is **false**
- `while` loops while the condition is **false**
- Functions are **defined** with `call` and **called** with `define`
- `true` is `false`, `false` is `true`
- `print` reads input, `input` prints output
- All error messages are **encouraging and positive**

## Quick Example

```worng
This line is ignored. So is this one.

// input "Hello, World!"
```

Output: `!dlroW ,olleH`

Strings are reversed on output by default. Use `~` for normal output:

```worng
// input ~"Hello, World!"
```

Output: `Hello, World!`

## Install

```bash
go install github.com/KashifKhn/worng/cmd/worng@latest
```

## Usage

```
worng run <file>      Run a .wrg file
worng run --repl      Interactive REPL
worng check <file>    Parse without running
worng fmt <file>      Format in-place
worng lsp             Start LSP server (stdio)
worng version         Print version
```

## File Extensions

`.wrg` (canonical), `.worng`, `.wrong`

## Documentation

- [Language Specification](docs/SPEC.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Roadmap](docs/ROADMAP.md)

## License

MIT
