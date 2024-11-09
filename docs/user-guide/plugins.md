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
Installing "foo" plugin...
Installed "foo" (source: github.com/org/tflint-ruleset-foo, version: 0.1.0)
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

When you install plugins with `tflint --init`, TFLint calls the GitHub API to get release metadata. By default, this is an unauthenticated request, subject to a rate limit of 60 requests per hour _per IP address_.

**Background:** [GitHub REST API: Rate Limiting](https://docs.github.com/en/rest/overview/resources-in-the-rest-api#rate-limiting)

If you fetch plugins frequently in CI, you may hit this rate limit. If you run TFLint in a shared CI environment such as GitHub Actions, you will share this quota with other tenants and may encounter rate limiting errors regardless of how often you run TFLint. 

To increase the rate limit, you can send an authenticated request by authenticating your requests with an access token, by setting the `GITHUB_TOKEN` environment variable. In GitHub Actions, you can pass the built-in `GITHUB_TOKEN` that is injected into each job.

It's also a good idea to cache the plugin directory, as TFLint will only send requests if plugins aren't installed. The [setup-tflint action](https://github.com/terraform-linters/setup-tflint#usage) includes an example of caching in GitHub Actions.

If you host your plugins on GitHub Enterprise Server (GHES), you may need to use a different token than on GitHub.com. In this case, you can use a host-specific token like `GITHUB_TOKEN_example_com`. The hostname must be normalized with Punycode. Use "_" instead of "." and "__" instead of "-".

```hcl
# GITHUB_TOKEN will be used
plugin "foo" {
  source = "github.com/org/tflint-ruleset-foo"
}

# GITHUB_TOKEN_example_com will be used preferentially and will fall back to GITHUB_TOKEN if not set.
plugin "bar" {
  source = "example.com/org/tflint-ruleset-bar"
}
```

You can reduce the usage of GitHub API with [`plugin_release_cache`](config.md#plugin_release_cache) and [`plugin_reduce_gh_api`](config.md#plugin_reduce_gh_api) in the configuration.

## Keeping plugins up to date

We recommend using automatic updates to keep your plugin version up-to-date. [Renovate supports TFLint plugins](https://docs.renovatebot.com/modules/manager/tflint-plugin/) to easily set up automated update workflows.

## Manual installation

You can also install the plugin manually. This is mainly useful for plugin development and for plugins that are not published on GitHub. In that case, omit the `source` and `version` attributes.

```hcl
plugin "foo" {
  enabled = true
}
```

When the plugin is enabled, TFLint invokes the `tflint-ruleset-[name]` (`tflint-ruleset-[name].exe` on Windows) binary in the plugin directory (For instance, `~/.tflint.d/plugins/tflint-ruleset-[name]`). So you should move the binary into the directory in advance.

## Bundled plugin

[TFLint Ruleset for Terraform Language](https://github.com/terraform-linters/tflint-ruleset-terraform) is built directly into TFLint binary. This is called a bundled plugin. Unlike other plugins, bundled plugins can be used without installation.

A bundled plugin is enabled by default without a plugin block declaration. The default config is below:

```hcl
plugin "terraform" {
  enabled = true
  preset  = "recommended"
}
```

You can also change the behavior of the bundled plugin by explicitly declaring a plugin block.

If you want to use a different version of tflint-ruleset-terraform instead of the bundled plugin, you can install it with `tflint --init` by specifying the `version` and `source`. In this case the bundled plugin will not be automatically enabled.

```hcl
plugin "terraform" {
  enabled = true
  preset  = "recommended"

  version = "0.1.0"
  source  = "github.com/terraform-linters/tflint-ruleset-terraform"
}
```

If you have tflint-ruleset-terraform manually installed, the bundled plugin will not be automatically enabled. In this case the manually installed version takes precedence.
