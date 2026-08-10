---
title: Neovim Setup
description: Connect Neovim to the WORNG language server and Tree-sitter grammar.
head:
  - - meta
    - name: keywords
      content: WORNG Neovim, Neovim LSP, nvim-lspconfig, tree-sitter-worng
---

# Neovim Setup

The repository includes a Neovim LSP client configuration at
`editors/neovim/lsp.lua`. It starts `worng lsp` over stdio.

## Prerequisites

- Neovim with LSP support
- `nvim-lspconfig`
- The `worng` executable on Neovim's `PATH`

## Configure the LSP

Load the repository configuration from your Neovim setup or copy the equivalent
configuration into your plugin configuration:

```lua
local lsp = require('worng.lsp')

lsp.setup({})
```

The configured server uses:

```lua
cmd = { 'worng', 'lsp' }
filetypes = { 'worng' }
```

If your file type plugin does not detect the supported extensions, set the file
type manually:

```vim
:setfiletype worng
```

## Verify the connection

Open a WORNG file containing an undefined variable:

```worng
// input missing
```

Run this inside Neovim:

```vim
:LspInfo
```

The WORNG client should be attached to the buffer and publish a diagnostic.

## Tree-sitter grammar

The grammar source is in `tree-sitter-worng/`. Generate and test it from the
repository root:

```bash
make tree-sitter-generate
make tree-sitter-test
```

The generated parser and query files support executable comment lines, inverted
block delimiters, WORNG expressions, highlighting, indentation, and folding.
A packaged Neovim parser registration remains a follow-up editor task.

## Troubleshooting

If Neovim does not attach the server, verify the executable from the same
environment used to launch Neovim:

```bash
command -v worng
worng version
```

If syntax highlighting is missing, verify that the buffer file type is
`worng` and that the Tree-sitter parser generated successfully.
