---
title: LSP Setup
description: Configure the WORNG Language Server for diagnostics, hover, completion, references, rename, formatting, and semantic tokens.
head:
  - - meta
    - name: keywords
      content: WORNG LSP, Language Server Protocol, worng lsp, editor diagnostics, WORNG autocomplete
---

# LSP Setup

WORNG includes a Language Server Protocol implementation in the `worng` CLI.
It communicates over standard input and output using JSON-RPC 2.0.

## Features

The server provides:

- Syntax and semantic diagnostics
- Encouraging diagnostic messages with source ranges and hints
- Keyword, variable, function, and `wronglib` completion
- Hover documentation for inverted WORNG behavior
- Go-to-definition
- References and rename
- Signature help
- Document symbols
- Semantic tokens
- Document formatting
- Full and incremental document synchronization

## Start the server

Verify that the binary is available:

```bash
worng version
```

Start the server with:

```bash
worng lsp
```

The command uses stdio. An LSP client must own the input and output streams.

## Generic client configuration

Configure an LSP client with:

```text
Command:    worng lsp
File types: .wrg, .worng, .wrong
Transport:  stdio
```

The client must use UTF-8, UTF-16, or UTF-32 position encoding supported by the
server. The server negotiates the encoding during `initialize`.

## Troubleshooting

Check that `worng` is on the editor process's `PATH`:

```bash
command -v worng
worng version
```

If diagnostics do not appear, verify that:

- The document uses `.wrg`, `.worng`, or `.wrong`.
- The client starts `worng lsp`, not `worng run`.
- The client sends `initialize` before document notifications.
- The editor's LSP log has no JSON-RPC framing error.

The JSON-RPC transport lives in `internal/jsonrpc`. LSP lifecycle, document
storage, indexing, diagnostics, and feature handlers live in `internal/lsp`.
The Go lexer and parser remain authoritative for language syntax.

See the [Neovim setup](/guide/neovim) for a concrete client configuration.
