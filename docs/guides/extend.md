# Extending TFLint

TFLint allows you to add your own rules via plugins. This can enforce organization-specific naming conventions and best practices.

Plugins are placed in the `~/.tflint.d/plugins` directory with the naming convention `tflint-ruleset-<NAME>.so`. You can explicitly enable the plugin by `.tflint.hcl` as follows:

```hcl
plugin "NAME" {
    enabled = true
}
```

That's all. Now you can freely add custom rules to TFLint!

Plugins are provided as a single `*.so` file and can be easily built with `go build --buildmode=plugin main.go`. You can see an example: https://github.com/terraform-linters/tflint-ruleset-template

The plugin must satisfy the following constraints:

- `Name` function is implemented under the `main` package.
  - It should return the plugin name as a string.
- `Version` function is implemented under the `main` package.
  - It should return the plugin version as a string.
- `NewRules` function is implemented under the `main` package.
  - It should return a list of rules that implement `plugin.Rule` interface.
- It is built using the same TFLint version.

Note the plugin package is only supported on Linux and macOS. Therefore, this plugin system does not work on Windows.
