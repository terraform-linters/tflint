# Editor Integration

TFLint can also act as a Language Server to integrate with various editors. This server conforms to [Language Server Protocol](https://microsoft.github.io/language-server-protocol/) v3.14.0 and can be used with an editor that implements any client. The following is an example in VS Code:

![demo](../assets/lsp_demo.gif)

This server can be started with the `--langserver` option:

```console
$ tflint --langserver
14:21:51 cli.go:185: Starting language server...
```

Currently, it only supports diagnostics and subscribes the following methods:

- `initialize`
- `initialized`
- `shutdown`
- `exit`
- `textDocument/didOpen`
- `textDocument/didClose`
- `textDocument/didChange`
- `workspace/didChangeWatchedFiles`
