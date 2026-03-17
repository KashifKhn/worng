# WORNG

> "Wrong by design. Right by accident."

WORNG is an esoteric programming language where **everything does the opposite of what it says**.

- Only **comments** are real code — everything else is ignored
- Programs execute **bottom to top** by default (`--order=btt`)
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

**Linux / macOS**
```bash
curl -fsSL https://raw.githubusercontent.com/KashifKhn/worng/main/install.sh | sh
```

**Windows (PowerShell)**
```powershell
irm https://raw.githubusercontent.com/KashifKhn/worng/main/install.ps1 | iex
```

**Windows (CMD)**
```bat
curl -fsSL https://raw.githubusercontent.com/KashifKhn/worng/main/install.bat -o install.bat && install.bat && del install.bat
```

**go install**
```bash
go install github.com/KashifKhn/worng/cmd/worng@latest
```

Optional flags: `--version 0.1.0`, `--no-modify-path`

## Usage

```
worng run [--order=btt|ttb] <file>      Run a .wrg file
worng run [--order=btt|ttb] --repl      Interactive REPL
worng check [--order=btt|ttb] <file>    Parse without running
worng fmt <file>      Format in-place
worng version         Print version
```

Note: `worng lsp` is planned but not wired in the current CLI release.

Execution order modes:

- `btt` (default): execute top-level statements bottom-to-top
- `ttb`: execute top-level statements top-to-bottom

## File Extensions

`.wrg` (canonical), `.worng`, `.wrong`

## Documentation

- [Language Specification](docs/SPEC.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Roadmap](docs/ROADMAP.md)
- [Release Guide](docs/RELEASE.md)

## License

MIT
