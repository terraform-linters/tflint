# Integración con editores de código

TFLint también puede actuar como servidor de lenguaje para integrarse con varios editores. Este servidor se ajusta al [Protocolo de servidor de lenguaje](https://microsoft.github.io/language-server-protocol/) v3.14.0 y puede utilizarse con un editor que implemente cualquier cliente. El siguiente es un ejemplo en VS Code:

![demo](../assets/lsp_demo.gif)

Este servidor puede iniciarse con el indicador `--langserver`:

```console
$ tflint --langserver
14:21:51 cli.go:185: Starting language server...
```

Actualmente, sólo admite diagnósticos y suscribe los siguientes métodos:

- `initialize`
- `initialized`
- `shutdown`
- `exit`
- `textDocument/didOpen`
- `textDocument/didClose`
- `textDocument/didChange`
- `workspace/didChangeWatchedFiles`
