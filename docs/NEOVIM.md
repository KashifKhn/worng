# Neovim Integration (WORNG First)

This project ships LSP inside the main CLI binary via `worng lsp`.

## Requirements

- Neovim 0.9+
- `nvim-lspconfig`
- `worng` binary on `PATH`

## Minimal setup

```lua
-- init.lua
local lspconfig = require('lspconfig')

lspconfig.worng_lsp = {
  default_config = {
    cmd = { 'worng', 'lsp' },
    filetypes = { 'worng' },
    root_dir = lspconfig.util.root_pattern('.git', 'go.mod'),
    single_file_support = true,
  },
}

lspconfig.worng_lsp.setup({})
```

## Filetype detection

```lua
vim.filetype.add({
  extension = {
    wrg = 'worng',
    worng = 'worng',
    wrong = 'worng',
  },
})
```

## Verify quickly

1. Open a `.wrg` file.
2. Run `:LspInfo` and verify `worng_lsp` is attached.
3. Test diagnostics with invalid snippet: `// if`.
4. Test hover on `if`.
5. Test completion after `wronglib.`

## Supported LSP methods

- diagnostics (`textDocument/publishDiagnostics`)
- hover
- completion
- definition
- references
- rename
- signatureHelp
- documentSymbol
- semanticTokens/full
- document formatting
