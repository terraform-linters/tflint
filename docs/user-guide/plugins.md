# Configuring Plugins

You can extend TFLint by installing any plugin. Declare plugins you want to use in the config file as follows:

```hcl
plugin "foo" {
  enabled = true
  version = "0.1.0"
  source  = "github.com/org/tflint-ruleset-foo"
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

### `signature`

Controls how TFLint verifies plugin releases. Valid values are:

- `auto`: Prefer `attestation` when available, then fall back to `pgp`. This is the default behavior.
- `attestation`: Require [artifact attestations](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds). Attestation verification in private repositories is not supported.
- `pgp`: Require PGP signature verification with `signing_key`.
- `none`: Skip plugin signature verification.

### `signing_key`

The signing key used when `signature = "pgp"` or `"auto"`.

## Plugin directory

Plugins are usually installed under `~/.tflint.d/plugins`. Exceptionally, if you already have `./.tflint.d/plugins` in your working directory, it will be installed there.

The automatically installed plugins are placed as `[plugin dir]/[source]/[version]/tflint-ruleset-[name]`. (`tflint-ruleset-[name].exe` in Windows).

If you want to change the plugin directory, you can change this with the [`plugin_dir`](config.md#plugin_dir) or `TFLINT_PLUGIN_DIR` environment variable.

## Dependency lockfile

`tflint --init` records the installed plugins in a dependency lockfile. Commit the lockfile to version control, just like Terraform's [dependency lock file](https://developer.hashicorp.com/terraform/language/files/dependency-lock). On later installations, TFLint verifies the plugins against the lockfile, so only the plugins that were trusted at the first installation can be installed. This mitigates supply chain attacks such as [repojacking](https://github.com/terraform-linters/tflint/issues/1634).

```console
$ tflint --init
Installing "foo" plugin...
Installed "foo" (source: github.com/org/tflint-ruleset-foo, version: 0.1.0)
Created lockfile .tflint.lock.hcl
```

### Lockfile location

The lockfile is placed next to the loaded config file. The name is derived from the config file name by replacing the `.hcl` or `.json` extension with `.lock.hcl`. This applies regardless of how the config file is found (`--config`, `TFLINT_CONFIG_FILE`, or the default lookup).

| Config file | Lockfile |
| --- | --- |
| `.tflint.hcl` | `.tflint.lock.hcl` |
| `.tflint.json` | `.tflint.lock.hcl` |
| `configs/aws.hcl` | `configs/aws.lock.hcl` |
| `~/.tflint.hcl` | `~/.tflint.lock.hcl` |

If no config file is found, no lockfile is used. When using `--chdir` or `--recursive`, the lockfile path is derived from the config file loaded in each directory.

### Lockfile format

```hcl
# This file is maintained automatically by "tflint --init".
# Manual edits may be lost in future updates.

lockfile_version = 1

plugin "github.com/org/tflint-ruleset-foo" {
  version       = "0.1.0"
  repository_id = 123456789
  hashes = {
    darwin_amd64  = "h1:0a1b2c..."
    linux_amd64   = "h1:3d4e5f..."
    windows_amd64 = "h1:6a7b8c..."
  }
}
```

- `lockfile_version`: The version of the lockfile format.
- `plugin` blocks: Locked plugins. The label is the plugin `source`.
  - `version`: The installed plugin version.
  - `repository_id`: The ID of the source repository on GitHub. The ID changes when the repository is deleted and recreated, even if it has the same name.
  - `hashes`: The hashes of the plugin binary per platform (`[GOOS]_[GOARCH]`). `h1:` is followed by the sha256 hash of the plugin binary (not the zip file). Note that this is different from Terraform's `h1:` hash.

If the plugin release contains the binary hashes for all platforms in `checksums.txt`, all of them are recorded. Otherwise, only the hash of the binary installed on the current platform is recorded, and hashes for other platforms are added when running `tflint --init` on those platforms.

If all the hashes are available in `checksums.txt`, every value in the lockfile can be obtained from the GitHub release and the GitHub API. This means that automated update workflows can update the lockfile without running `tflint --init`.

### Installation behavior

When installing a plugin, `tflint --init` behaves as follows:

- If the plugin is not recorded in the lockfile, the plugin is installed, and a new entry is added to the lockfile.
- If the plugin is recorded with the same version, the repository ID and hashes must match the lockfile. Otherwise, the installation fails and nothing is installed. A hash for a platform not yet recorded is added to the lockfile.
- If the plugin is recorded with a different version, the plugin is installed, and the hashes are replaced. However, if the repository ID does not match, the installation fails. If the repository change is expected, remove the entry from the lockfile and run `tflint --init` again.

If the lockfile exists, all plugins must be recorded in the lockfile with the declared version. Plugins that are already installed are skipped if they are recorded with the declared version. Otherwise, they are reinstalled to record them in the lockfile.

If the lockfile does not exist, plugins that are already installed are skipped as before. The lockfile is created only when a new plugin is installed. At that time, plugins that are already installed are also reinstalled to record them in the new lockfile.

Entries for plugins no longer declared in the config file are removed from the lockfile. Plugins without `source` and `version` (manual installation and the bundled plugin) are not recorded in the lockfile.

When launching plugins, if the lockfile exists, TFLint (including the language server) checks the installed plugins against it. If the plugin is not recorded in the lockfile, or the plugin version does not match the lockfile, an error is returned. In that case, run `tflint --init` to update the lockfile.

### Lockfile versioning

The `lockfile_version` is bumped only for changes that older TFLint cannot safely ignore, such as structural changes or additions related to verification. Attributes that older TFLint can safely ignore may be added without bumping the version, and unknown attributes are ignored when reading. When rewriting the lockfile, TFLint does not bump the version of an existing lockfile unless necessary.

If the lockfile version is newer than the version supported by TFLint, TFLint returns an error. In that case, upgrade TFLint.

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

## Keeping plugins up to date

We recommend using automatic updates to keep your plugin version up-to-date. [Renovate supports TFLint plugins](https://docs.renovatebot.com/modules/manager/tflint-plugin/) to easily set up automated update workflows.

## Manual installation

You can also install the plugin manually. This is mainly useful for plugin development and for plugins that are not published on GitHub. In that case, omit the `source` and `version` attributes.

```hcl
plugin "foo" {
  enabled = true
}
```

When enabled, TFLint looks for the plugin binary following the pattern tflint-ruleset-\[name] (or tflint-ruleset-\[name].exe on Windows) directly in your plugins folder (e.g., ~/.tflint.d/plugins/ on Unix-like systems or C:\Users\\[username]\\.tflint.d\plugins\ on Windows).

Thus, with the configuration above where the plugin name is "foo", the executable must be named tflint-ruleset-foo (or tflint-ruleset-foo.exe on Windows). So you should move the binary into the plugin directory in advance.

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
