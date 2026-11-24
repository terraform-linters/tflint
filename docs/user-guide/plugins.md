# Configuring Plugins

You can extend TFLint with plugins published on GitHub or GitLab. Declare the plugins you want to use in the config file.

For GitHub.com or GitHub Enterprise Server:

```hcl
plugin "foo" {
  enabled = true
  version = "0.1.0"
  source  = "github.com/org/tflint-ruleset-foo"
}
```

For GitLab.com or self-managed GitLab, also set `source_type = "gitlab"`:

```hcl
plugin "foo" {
  enabled     = true
  version     = "0.1.0"
  source      = "gitlab.example.com/group/subgroup/tflint-ruleset-foo"
  source_type = "gitlab"
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

The repository to install the plugin from, without a URL scheme or `.git` suffix.

| Release host | Source format | Example |
| --- | --- | --- |
| GitHub | `${host}/${owner}/${repo}` | `github.com/org/tflint-ruleset-foo` |
| GitLab | `${host}/${namespace}/${repo}` | `gitlab.com/group/subgroup/tflint-ruleset-foo` |

The host can also be a GitHub Enterprise Server or self-managed GitLab instance. GitLab namespaces can include multiple subgroups.

### `source_type`

The release API to use: `github` (default) or `gitlab`. Set this to `gitlab` for both GitLab.com and self-managed GitLab. TFLint does not infer the source type from the hostname, so omitting this attribute uses the GitHub API even for a GitLab host.

### `version`

Plugin version. Do not prefix with "v". This attribute cannot be omitted when the `source` is set. Version constraints (like `>= 0.3`) are not supported.

### `signature`

Controls how TFLint verifies plugin releases. Valid values are:

- `auto`: Prefer `attestation` when available, then fall back to `pgp`. This is the default behavior.
- `attestation`: Require [artifact attestations](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds). This is supported for public GitHub.com repositories. Private repositories, GitHub Enterprise Server, and GitLab sources do not support attestation verification.
- `pgp`: Require PGP signature verification with `signing_key`.
- `none`: Skip plugin signature verification.

SHA-256 checksums are verified in every mode, including `none`. For GitLab sources, `auto` uses PGP when a `signing_key` is configured; otherwise, installation succeeds with a warning that the signature was not verified. See the [publishing requirements](../developer-guide/plugins.md#4-publishing-a-release) for the release files used by both hosts.

### `signing_key`

The signing key used when `signature = "pgp"` or `"auto"`.

## Plugin directory

Plugins are usually installed under `~/.tflint.d/plugins`. Exceptionally, if you already have `./.tflint.d/plugins` in your working directory, it will be installed there.

The automatically installed plugins are placed as `[plugin dir]/[source]/[version]/tflint-ruleset-[name]`. (`tflint-ruleset-[name].exe` in Windows).

If you want to change the plugin directory, you can change this with the [`plugin_dir`](config.md#plugin_dir) or `TFLINT_PLUGIN_DIR` environment variable.

## Authentication and rate limits

### GitHub

When you install plugins from GitHub with `tflint --init`, TFLint calls the GitHub API to get release metadata. By default, this is an unauthenticated request, subject to a rate limit of 60 requests per hour _per IP address_.

**Background:** [GitHub REST API: Rate limits](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api)

If you fetch plugins frequently in CI, you may hit this rate limit. If you run TFLint in a shared CI environment such as GitHub Actions, you will share this quota with other tenants and may encounter rate limiting errors regardless of how often you run TFLint.

To increase the rate limit, you can send an authenticated request by authenticating your requests with an access token, by setting the `GITHUB_TOKEN` environment variable. In GitHub Actions, you can pass the built-in `GITHUB_TOKEN` that is injected into each job.

If you host your plugins on GitHub Enterprise Server (GHES), you may need to use a different token than on GitHub.com. In this case, you can use a host-specific token like `GITHUB_TOKEN_example_com`. The hostname must be normalized with Punycode. Use "_" instead of "." and "__" instead of "-".

```hcl
# GITHUB_TOKEN will be used
plugin "foo" {
  version = "0.1.0"
  source  = "github.com/org/tflint-ruleset-foo"
}

# GITHUB_TOKEN_example_com will be used preferentially and will fall back to GITHUB_TOKEN if not set.
plugin "bar" {
  version = "0.1.0"
  source  = "example.com/org/tflint-ruleset-bar"
}
```

### GitLab

TFLint uses the GitLab API at `/api/v4` to find [release assets](https://docs.gitlab.com/user/project/releases/release_fields/#release-assets). Public projects can be accessed without a token. Both the API and asset URLs must use HTTPS.

For private projects, TFLint selects the first non-empty credential in this order:

1. `GITLAB_TOKEN_${source_host}`, using the same hostname encoding as `GITHUB_TOKEN_${source_host}`.
2. `GITLAB_TOKEN`.
3. `CI_JOB_TOKEN`, only when `CI_SERVER_URL` has the same HTTPS host and port as the plugin source.

Access tokens are sent as `PRIVATE-TOKEN`, and CI job tokens as `JOB-TOKEN`. The token must be permitted to read both the release and its assets. Cross-project CI downloads may require the source project to add the consuming project or group to its [job token allowlist](https://docs.gitlab.com/ci/jobs/ci_job_token/).

Credentials are only sent to the plugin source's exact HTTPS origin. Asset links and redirects to another host or port are followed without the GitLab token, so external assets must be accessible without it.

### Caching plugins

For either source type, caching the plugin directory avoids API requests for plugins that are already installed. The [setup-tflint action](https://github.com/terraform-linters/setup-tflint#usage) includes a GitHub Actions example; GitLab CI jobs can cache the same directory.

## Keeping plugins up to date

We recommend using automatic updates to keep your plugin version up-to-date. [Renovate supports TFLint plugins](https://docs.renovatebot.com/modules/manager/tflint-plugin/) to easily set up automated update workflows.

## Manual installation

You can also install the plugin manually. This is mainly useful for plugin development and for plugins that are not published on GitHub or GitLab. In that case, omit the `source` and `version` attributes.

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
