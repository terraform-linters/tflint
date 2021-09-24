# TFLint
[![Build Status](https://github.com/terraform-linters/tflint/workflows/build/badge.svg?branch=master)](https://github.com/terraform-linters/tflint/actions)
[![GitHub release](https://img.shields.io/github/release/terraform-linters/tflint.svg)](https://github.com/terraform-linters/tflint/releases/latest)
[![Terraform Compatibility](https://img.shields.io/badge/terraform-%3E%3D%200.12-blue)](docs/user-guide/compatibility.md)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/terraform-linters/tflint)](https://goreportcard.com/report/github.com/terraform-linters/tflint)
[![Homebrew](https://img.shields.io/badge/dynamic/json.svg?url=https://formulae.brew.sh/api/formula/tflint.json&query=$.versions.stable&label=homebrew)](https://formulae.brew.sh/formula/tflint)

A Pluggable [Terraform](https://www.terraform.io/) Linter

## Features

TFLint is a framework and each feature is provided by plugins, the key features are as follows:

- Find possible errors (like illegal instance types) for Major Cloud providers (AWS/Azure/GCP).
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

### Docker

Instead of installing directly, you can use the Docker images:

| Name | Description |
| ---- | ----------- |
| [ghcr.io/terraform-linters/tflint](https://github.com/terraform-linters/tflint/pkgs/container/tflint) | Basic image |
| [ghcr.io/terraform-linters/tflint-bundle](https://github.com/terraform-linters/tflint-bundle/pkgs/container/tflint-bundle) | A Docker image with TFLint and ruleset plugins |

Example:

```console
docker run --rm -v $(pwd):/data -t ghcr.io/terraform-linters/tflint
```

### GitHub Actions

If you want to run on GitHub Actions, [setup-tflint](https://github.com/terraform-linters/setup-tflint) action is available.

## Getting Started

If you are using an AWS/Azure/GCP provider, it is a good idea to install the plugin and try it according to each usage:

- [Amazon Web Services](https://github.com/terraform-linters/tflint-ruleset-aws)
- [Microsoft Azure](https://github.com/terraform-linters/tflint-ruleset-azurerm)
- [Google Cloud Platform](https://github.com/terraform-linters/tflint-ruleset-google)

Rules for the Terraform Language is built into the TFLint binary, so you don't need to install any plugins. Please see [Rules](docs/rules) for a list of available rules.

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

## Usage

TFLint inspects files under the current directory by default. You can change the behavior with the following options/arguments:

```
$ tflint --help
Usage:
  tflint [OPTIONS] [FILE or DIR...]

Application Options:
  -v, --version                                                 Print TFLint version
      --init                                                    Install plugins
      --langserver                                              Start language server
  -f, --format=[default|json|checkstyle|junit|compact|sarif]    Output format (default: default)
  -c, --config=FILE                                             Config file name (default: .tflint.hcl)
      --ignore-module=SOURCE                                    Ignore module sources
      --enable-rule=RULE_NAME                                   Enable rules from the command line
      --disable-rule=RULE_NAME                                  Disable rules from the command line
      --only=RULE_NAME                                          Enable only this rule, disabling all other defaults. Can be specified multiple times
      --enable-plugin=PLUGIN_NAME                               Enable plugins from the command line
      --var-file=FILE                                           Terraform variable file name
      --var='foo=bar'                                           Set a Terraform variable
      --module                                                  Inspect modules
      --force                                                   Return zero exit status even if issues found
      --no-color                                                Disable colorized output
      --loglevel=[trace|debug|info|warn|error]                  Change the loglevel

Help Options:
  -h, --help                                                    Show this help message

```

See [User Guide](docs/user-guide) for details.

## FAQ

### Does TFLint check modules recursively?
No. TFLint always checks only the current root module (no recursive check). However, you can check calling child modules based on module arguments by enabling [Module Inspection](docs/user-guide/module-inspection.md). This allows you to check that you are not passing illegal values to the module.

Note that if you want to recursively inspect local modules, you need to run them in each directory. This is a limitation that occurs because Terraform always works for one directory. TFLint tries to emulate Terraform's semantics, so cannot perform recursive inspection.

### Do I need to install Terraform for TFLint to work?
No. TFLint works as a single binary because Terraform is embedded as a library. Note that this means that the version of Terraform used is determined for each TFLint version. See also [Compatibility with Terraform](docs/user-guide/compatibility.md).

### TFLint reports a loading error in my code, but this is valid in Terraform. Why?
First, check the version of Terraform and TFLint you are using. TFLint loads files differently than the installed Terraform, so an error can occur if the version of Terraform supported by TFLint is different from the installed Terraform.

## Debugging

If you don't get the expected behavior, you can see the detailed logs when running with `TFLINT_LOG` environment variable.

```console
$ TFLINT_LOG=debug tflint
```

## Developing

See [Developer Guide](docs/developer-guide).

## Stargazers over time

[![Stargazers over time](https://starchart.cc/terraform-linters/tflint.svg)](https://starchart.cc/terraform-linters/tflint)
