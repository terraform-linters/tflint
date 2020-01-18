# Extending TFLint

TFLint allows you to add your own rules via plugins. This can enforce organization-specific naming conventions and best practices.

Plugins are placed in the `~/.tflint.d/plugins` (or `./tflint.d/plugins`) directory with the naming convention `tflint-ruleset-<NAME>` (`tflint-ruleset-<NAME>.exe` on Windows). You can explicitly enable the plugin by `.tflint.hcl` as follows:

```hcl
plugin "NAME" {
    enabled = true
}
```

That's all. Now you can freely add custom rules to TFLint!

A plugins is provided as a single binary and can be built using [`tflint-plugin-sdk`](https://github.com/terraform-linters/tflint-plugin-sdk). If you are interested in writing plugins, please see here.
