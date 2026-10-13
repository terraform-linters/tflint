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

The repository to install the plugin from, without a URL scheme or `.git` suffix. For GitHub, use `${host}/${owner}/${repo}`, such as `github.com/org/repo`. For GitLab, use `${host}/${namespace}/${repo}`; the namespace can include subgroups.

### `source_type`

The release API to use: `github` (default) or `gitlab`. Set this to `gitlab` for both GitLab.com and self-managed GitLab. Other hosts continue to use the GitHub Enterprise API unless this attribute is set.

### `version`

Plugin version. Do not prefix with "v". This attribute cannot be omitted when the `source` is set. Version constraints (like `>= 0.3`) are not supported.

### `signature`

Controls how TFLint verifies plugin releases. Valid values are:

- `auto`: Prefer `attestation` when available, then fall back to `pgp`. This is the default behavior.
- `attestation`: Require [artifact attestations](https://docs.github.com/en/actions/security-for-github-actions/using-artifact-attestations/using-artifact-attestations-to-establish-provenance-for-builds). Attestation verification in private repositories and GitLab sources is not supported.
- `pgp`: Require PGP signature verification with `signing_key`.
- `none`: Skip plugin signature verification.

### `signing_key`

The signing key used when `signature = "pgp"` or `"auto"`.

## Plugin directory

Plugins are usually installed under `~/.tflint.d/plugins`. Exceptionally, if you already have `./.tflint.d/plugins` in your working directory, it will be installed there.

The automatically installed plugins are placed as `[plugin dir]/[source]/[version]/tflint-ruleset-[name]`. (`tflint-ruleset-[name].exe` in Windows).

If you want to change the plugin directory, you can change this with the [`plugin_dir`](config.md#plugin_dir) or `TFLINT_PLUGIN_DIR` environment variable.

## Avoiding rate limiting

When you install plugins from GitHub with `tflint --init`, TFLint calls the GitHub API to get release metadata. By default, this is an unauthenticated request, subject to a rate limit of 60 requests per hour _per IP address_.

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

## GitLab Releases

Plugins published as [GitLab release assets](https://docs.gitlab.com/user/project/releases/release_fields/#release-assets) can be installed with `tflint --init`:

```hcl
plugin "example" {
  enabled     = true
  source_type = "gitlab"
  source      = "gitlab.example.com/group/subgroup/tflint-ruleset-example"
  version     = "0.1.0"
}
```

TFLint requests the `v0.1.0` release through the GitLab API at `/api/v4`. The release must contain asset links whose names match the archive (`tflint-ruleset-example_${GOOS}_${GOARCH}.zip`) and `checksums.txt`. Each link's URL must point to the downloadable file, for example in the Generic Package Registry. Both the GitLab API and asset URLs must use HTTPS.

SHA-256 checksums are always verified. To verify the publisher as well, configure `signing_key` and publish a `checksums.txt.sig` asset using the same PGP signing format as GitHub releases. `signature = "pgp"` requires verification; `auto` uses PGP when a key is configured and otherwise installs with the existing unverified-signature warning. GitHub artifact attestations cannot be used for GitLab sources.

For private projects, TFLint selects credentials in this order:

1. `GITLAB_TOKEN_${source_host}`, using the same hostname encoding as `GITHUB_TOKEN_${source_host}`.
2. `GITLAB_TOKEN`.
3. `CI_JOB_TOKEN`, only when `CI_SERVER_URL` has the same HTTPS host and port as the plugin source.

Access tokens are sent as `PRIVATE-TOKEN`, and CI job tokens as `JOB-TOKEN`. The token must be permitted to read both the release and its assets. Cross-project CI downloads may require the source project to add the consuming project or group to its [job token allowlist](https://docs.gitlab.com/ci/jobs/ci_job_token/).

Credentials are only sent to the plugin source's exact HTTPS origin. Asset links and redirects to another host or port are followed without the GitLab token, so external assets must be accessible without it.

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
