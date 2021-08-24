# Redis language server

Allow autocompletion, command execution, documentation for Redis using the Language Server Protocol.

### Supported messages

- [x] Autocompletion (```textDocument/completion```)
- [x] Documentation (```completionItem/resolve```)
- [x] Execute Redis commands (```workspace/executeCommand```)
- [ ] Hover (```textDocument/hover```)
- [ ] Reflect configuration changes in server (```workspace/didChangeConfiguration``` and ```workspace/configuration```)

### Installation

If you have Go installed:

```bash
go get github.com/fagnercarvalho/redis-lsp
```

Or check the [Releases](https://github.com/fagnercarvalho/redis-lsp/releases) page.

### Inspiration 

- [sqls](https://github.com/lighttiger2505/sqls).
- [gopls](https://github.com/golang/tools/tree/master/gopls).