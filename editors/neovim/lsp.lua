local lspconfig = require('lspconfig')

lspconfig.worng_lsp = {
  default_config = {
    cmd = { 'worng', 'lsp' },
    filetypes = { 'worng' },
    root_dir = lspconfig.util.root_pattern('.git', 'go.mod'),
    single_file_support = true,
  },
}

return {
  setup = function(opts)
    lspconfig.worng_lsp.setup(opts or {})
  end,
}
