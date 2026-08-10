# WORNG LSP Architecture

## Overview

WORNG LSP runs inside `worng lsp` and uses stdio JSON-RPC.

Core modules:

- `internal/jsonrpc`: framing + JSON-RPC transport
- `internal/lsp/server.go`: lifecycle and method dispatch
- `internal/lsp/diagnostics.go`: analysis scheduling + diagnostics publish
- `internal/lsp/features_*`: language features (hover/completion/definition/references/rename/signature/symbols/semantic tokens)
- `internal/lsp/index.go`: workspace symbol index
- `internal/lsp/incremental.go`: incremental document update helpers

## Supported Features

- initialize / shutdown / exit lifecycle
- full and incremental text sync handling
- publish diagnostics
- hover
- completion
- definition
- references
- rename
- signature help
- document symbols
- semantic tokens full
- document formatting

## Extending LSP

1. Add typed request/response models in `internal/lsp/lsproto/types_generated.go`.
2. Add handler branch in `internal/lsp/server.go`.
3. Implement feature logic in `internal/lsp/features_*.go`.
4. Add tests (unit + transcript style) in `internal/lsp/*_test.go`.

## Coverage Gates

CI enforces package coverage floors:

- `internal/jsonrpc >= 95%`
- `internal/lsp >= 95%`
