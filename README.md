# TFLint
[![Build Status](https://github.com/terraform-linters/tflint/actions/workflows/build.yml/badge.svg?branch=master)](https://github.com/terraform-linters/tflint/actions)
[![GitHub release](https://img.shields.io/github/release/terraform-linters/tflint.svg)](https://github.com/terraform-linters/tflint/releases/latest)
[![Terraform Compatibility](https://img.shields.io/badge/terraform-%3E%3D%201.0-blue)](docs/user-guide/compatibility.md)
[![License: MPL 2.0 + BUSL 1.1](https://img.shields.io/badge/License-MPL%202.0%20+%20BUSL%201.1-blue.svg)](#license)
[![Go Report Card](https://goreportcard.com/badge/github.com/terraform-linters/tflint)](https://goreportcard.com/report/github.com/terraform-linters/tflint)
[![Homebrew](https://img.shields.io/badge/dynamic/json.svg?url=https://formulae.brew.sh/api/formula/tflint.json&query=$.versions.stable&label=homebrew)](https://formulae.brew.sh/formula/tflint)

A Pluggable [Terraform](https://www.terraform.io/) Linter

## Features

TFLint is a framework and each feature is provided by plugins, the key features are as follows:

- Find possible errors (like invalid instance types) for Major Cloud providers (AWS/Azure/GCP).
- Warn about deprecated syntax, unused declarations.
- Enforce best practices, naming conventions.

## Installation

Bash script (Linux):

```console
curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh | bash
```

Homebrew (macOS):

```console
brew install tflint
```

Chocolatey (Windows):

```cmd
choco install tflint
```

NOTE: The Chocolatey package is NOT directly maintained by the TFLint maintainers. The latest version is always available by manual installation.

### Verification

#### GitHub CLI (Recommended)

[Artifact Attestations](https://docs.github.com/en/actions/security-guides/using-artifact-attestations-to-establish-provenance-for-builds) are available that can be verified using the GitHub CLI.

```console
gh attestation verify checksums.txt -R terraform-linters/tflint
sha256sum --ignore-missing -c checksums.txt
```

#### Cosign

[Cosign](https://github.com/sigstore/cosign) `verify-blob` command ensures that the release was built with GitHub Actions in this repository.

```console
cosign verify-blob --certificate=checksums.txt.pem --signature=checksums.txt.keyless.sig --certificate-identity-regexp="^https://github.com/terraform-linters/tflint" --certificate-oidc-issuer=https://token.actions.githubusercontent.com checksums.txt
sha256sum --ignore-missing -c checksums.txt
```

### Docker

Instead of installing directly, you can use the Docker image:

```console
docker run --rm -v $(pwd):/data -t ghcr.io/terraform-linters/tflint
```

To download plugins, you can override the entrypoint to a shell (`sh`) to run `--init` and the main command in a single `docker run` command:

```console
 docker run --rm -v $(pwd):/data -t --entrypoint /bin/sh ghcr.io/terraform-linters/tflint -c "tflint --init && tflint"
```

### GitHub Actions

If you want to run on GitHub Actions, [setup-tflint](https://github.com/terraform-linters/setup-tflint) action is available.

## Getting Started

First, enable rules for [Terraform Language](https://www.terraform.io/language) (e.g. warn about deprecated syntax, unused declarations). [TFLint Ruleset for Terraform Language](https://github.com/terraform-linters/tflint-ruleset-terraform) is bundled with TFLint, so you can use it without installing it separately.

The bundled plugin enables the "recommended" preset by default, but you can disable the plugin or use a different preset. Declare the plugin block in `.tflint.hcl` like this:

```hcl
plugin "terraform" {
  enabled = true
  preset  = "recommended"
}
```

See the [tflint-ruleset-terraform documentation](https://github.com/terraform-linters/tflint-ruleset-terraform/blob/main/docs/configuration.md) for more information.

Next, If you are using an AWS/Azure/GCP provider, it is a good idea to install the plugin and try it according to each usage:

- [Amazon Web Services](https://github.com/terraform-linters/tflint-ruleset-aws)
- [Microsoft Azure](https://github.com/terraform-linters/tflint-ruleset-azurerm)
- [Google Cloud Platform](https://github.com/terraform-linters/tflint-ruleset-google)

If you want to extend TFLint with other plugins, you can declare the plugins in the config file and easily install them with `tflint --init`.

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

See also [Configuring Plugins](docs/user-guide/plugins.md).

You can discover plugins from other organizations on GitHub via the [`tflint-ruleset`](https://github.com/topics/tflint-ruleset) topic.

If you want to add custom rules that are not in existing plugins, you can build your own plugin or write your own policy in Rego. See [Writing Plugins](docs/developer-guide/plugins.md) or [OPA Ruleset](https://github.com/terraform-linters/tflint-ruleset-opa).

## Usage

TFLint inspects files under the current directory by default. You can change the behavior with the following options/arguments:

```
$ tflint --help
Usage:
  tflint --chdir=DIR/--recursive [OPTIONS]

Application Options:
  -v, --version                                                 Print TFLint version
      --init                                                    Install plugins
      --langserver                                              Start language server
  -f, --format=[default|json|checkstyle|junit|compact|sarif]    Output format
  -c, --config=FILE                                             Config file name (default: .tflint.hcl)
      --ignore-module=SOURCE                                    Ignore module sources
      --enable-rule=RULE_NAME                                   Enable rules from the command line
      --disable-rule=RULE_NAME                                  Disable rules from the command line
      --only=RULE_NAME                                          Enable only this rule, disabling all other defaults. Can be specified multiple times
      --enable-plugin=PLUGIN_NAME                               Enable plugins from the command line
      --var-file=FILE                                           Terraform variable file name
      --var='foo=bar'                                           Set a Terraform variable
      --call-module-type=[all|local|none]                       Types of module to call (default: local)
      --chdir=DIR                                               Switch to a different working directory before executing the command
      --recursive                                               Run command in each directory recursively
      --filter=FILE                                             Filter issues by file names or globs
      --force                                                   Return zero exit status even if issues found
      --minimum-failure-severity=[error|warning|notice]         Sets minimum severity level for exiting with a non-zero error code
      --color                                                   Enable colorized output
      --no-color                                                Disable colorized output
      --fix                                                     Fix issues automatically
      --no-parallel-runners                                     Disable per-runner parallelism
      --max-workers=N                                           Set maximum number of workers in recursive inspection (default: number of CPUs)

Help Options:
  -h, --help                                                    Show this help message
```

See [User Guide](docs/user-guide) for details.

## Debugging

If you don't get the expected behavior, you can see the detailed logs when running with `TFLINT_LOG` environment variable.

```console
$ TFLINT_LOG=debug tflint
```

## Developing

See [Developer Guide](docs/developer-guide).

## Security

If you find a security vulnerability, please refer our [security policy](SECURITY.md).

## License

Please note that although much of this project is licensed under MPL 2.0, some files in the `terraform` package are licensed under BUSL 1.1.

For the reasons stated above, the executable forms (release binaries) is bound by both licenses.

See also https://discuss.hashicorp.com/t/hashicorp-projects-changing-license-to-business-source-license-v1-1/57106/7

## Stargazers over time

[![Stargazers over time](https://starchart.cc/terraform-linters/tflint.svg)](https://starchart.cc/terraform-linters/tflint)
