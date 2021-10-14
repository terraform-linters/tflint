# Configuring Plugins

You can extend TFLint by installing any plugin. Declare plugins you want to use in the config file as follows:

```hcl
plugin "foo" {
  enabled = true
  version = "0.1.0"
  source  = "github.com/org/tflint-ruleset-foo"

  signing_key = <<-KEY
  -----BEGIN PGP PUBLIC KEY BLOCK-----

  mQINBFzpPOMBEADOat4P4z0jvXaYdhfy+UcGivb2XYgGSPQycTgeW1YuGLYdfrwz
  9okJj9pMMWgt/HpW8WrJOLv7fGecFT3eIVGDOzyT8j2GIRJdXjv8ZbZIn1Q+1V72
  AkqlyThflWOZf8GFrOw+UAR1OASzR00EDxC9BqWtW5YZYfwFUQnmhxU+9Cd92e6i
  ...
  KEY
}
```

After declaring the `version` and `source`, `tflint --init` can automatically install the plugin.

```console
$ tflint --init
Installing `foo` plugin...
Installed `foo` (source: github.com/org/tflint-ruleset-foo, version: 0.1.0)
$ tflint -v
TFLint version 0.28.1
+ ruleset.foo (0.1.0)
```

See also [Configuring TFLint](config.md) for the config file schema.

## Attributes

This section describes the attributes reserved by TFLint. Except for these, each plugin can extend the schema by defining any attributes/blocks. See the documentation for each plugin for details.

### `enabled` (required)

Enable the plugin. If set to false, the rules will not be used even if the plugin is installed.

### `source`

The source URL to install the plugin. Must be in the format `github.com/org/repo`.

### `version`

Plugin version. Do not prefix with "v". This attribute cannot be omitted when the `source` is set. Version constraints (like `>= 0.3`) are not supported.

### `signing_key`

Plugin developer's PGP public signing key. When this attribute is set, TFLint will automatically verify the signature of the checksum file downloaded from GitHub. It is recommended to set it to prevent supply chain attacks.

Plugins under the terraform-linters organization (AWS/GCP/Azure ruleset plugins) can use the built-in signing key, so this attribute can be omitted.

## Plugin directory

Plugins are usually installed under `~/.tflint.d/plugins`. Exceptionally, if you already have `./.tflint.d/plugins` in your working directory, it will be installed there.

The automatically installed plugins are placed as `[plugin dir]/[source]/[version]/tflint-ruleset-[name]`. (`tflint-ruleset-[name].exe` in Windows).

If you want to change the plugin directory, you can change this with the [`plugin_dir`](config.md#plugin_dir) or `TFLINT_PLUGIN_DIR` environment variable.

## Avoiding rate limiting

When you install plugins with `tflint --init`, call the GitHub API to get release metadata. This is typically an unauthenticated request with a rate limit of 60 requests per hour.

https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting

This limitation can be a problem if you need to run `--init` frequently, such as in CI environments. If you want to increase the rate limit, you can send an authenticated request by setting an OAuth2 access token in the `GITHUB_TOKEN` environment variable.

It's also a good idea to cache the plugin directory, as TFLint will only send requests if plugins aren't installed. See also the [setup-tflint's example](https://github.com/terraform-linters/setup-tflint#usage).

## Manual installation

You can also install the plugin manually. This is mainly useful for plugin development and for plugins that are not published on GitHub. In that case, omit the `source` and `version` attributes.

```hcl
plugin "foo" {
  enabled = true
}
```

When the plugin is enabled, TFLint invokes the `tflint-ruleset-[name]` (`tflint-ruleset-[name].exe` on Windows) binary in the plugin directory (For instance, `~/.tflint.d/plugins/tflint-ruleset-[name]`). So you should move the binary into the directory in advance.
