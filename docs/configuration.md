# Configuration

This plugin can take advantage of additional features by configuring the plugin block. Currently, this configuration is only available for preset.

Here's an example:

```hcl
plugin "terraform" {
    // Plugin common attributes

    preset = "recommended"
}
```

## `preset`

Default: `all` (`recommended` for the bundled plugin)

Enable multiple rules at once. Please see [Rules](rules/README.md) for details. Possible values are `recommended` and `all`.

When using the bundled plugin built into TFLint, you can use this plugin without declaring a "plugin" block. In this case the default is `recommeneded`.
