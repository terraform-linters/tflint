# TFLint
[![Build Status](https://github.com/terraform-linters/tflint/workflows/build/badge.svg?branch=master)](https://github.com/terraform-linters/tflint/actions)
[![GitHub release](https://img.shields.io/github/release/terraform-linters/tflint.svg)](https://github.com/terraform-linters/tflint/releases/latest)
[![Terraform Compatibility](https://img.shields.io/badge/terraform-%3E%3D%200.12-blue)](docs/guides/compatibility.md)
[![Docker Hub](https://img.shields.io/badge/docker-ready-blue.svg)](https://hub.docker.com/r/wata727/tflint/)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/terraform-linters/tflint)](https://goreportcard.com/report/github.com/terraform-linters/tflint)

TFLint is a [Terraform](https://www.terraform.io/) linter focused on possible errors, best practices, etc.

## Why TFLint is required?

Terraform is a great tool for Infrastructure as Code. However, many of these tools don't validate provider-specific issues. For example, see the following configuration file:

```hcl
resource "aws_instance" "foo" {
  ami           = "ami-0ff8a91507f77f867"
  instance_type = "t1.2xlarge" # invalid type!
}
```

Since `t1.2xlarge` is a nonexistent instance type, an error will occur when you run `terraform apply`. But `terraform plan` and `terraform validate` cannot find this possible error beforehand. That's because it's an AWS provider-specific issue and it's valid as a Terraform configuration.

TFLint finds such errors in advance:

![demo](docs/assets/demo.gif)

## Installation

You can download the binary built for your architecture from [the latest release](https://github.com/terraform-linters/tflint/releases/latest). The following is an example of installation on macOS:

```console
$ curl --location https://github.com/terraform-linters/tflint/releases/download/v0.22.0/tflint_darwin_amd64.zip --output tflint_darwin_amd64.zip
$ unzip tflint_darwin_amd64.zip
Archive:  tflint_darwin_amd64.zip
  inflating: tflint
$ install tflint /usr/local/bin
$ tflint -v
```

For Linux based OS, you can use the [`install_linux.sh`](https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh) to automate the installation process, or try the following oneliner to download latest binary for AMD64 architecture.
```
$ curl -L "$(curl -Ls https://api.github.com/repos/terraform-linters/tflint/releases/latest | grep -o -E "https://.+?_linux_amd64.zip")" -o tflint.zip && unzip tflint.zip && rm tflint.zip
```

### Homebrew

macOS users can also use [Homebrew](https://brew.sh) to install TFLint:

```console
$ brew install tflint
```

### Chocolatey

Windows users can use [Chocolatey](https://chocolatey.org):

```cmd
choco install tflint
```

### Docker

You can also use [TFLint via Docker](https://hub.docker.com/r/wata727/tflint/).

```console
$ docker run --rm -v $(pwd):/data -t wata727/tflint
```

## Features

700+ rules are available. See [Rules](docs/rules).

## Providers

TFLint supports multiple providers via plugins. The following is the Major Cloud support status.

|name|status|description|
|---|---|---|
|[AWS](https://github.com/terraform-linters/tflint-ruleset-aws)|Available|Inspections for AWS resources are now built into TFLint. So, it is not necessary to install the plugin separately. In the future, these will be cut out to the plugin, but all are in progress.|
|[Azure](https://github.com/terraform-linters/tflint-ruleset-azurerm)|Experimental|Experimental support has been started. You can inspect Azure resources by installing the plugin.|
|[Google Cloud Platform](https://github.com/terraform-linters/tflint-ruleset-google)|Experimental|Experimental support has been started. You can inspect GCP resources by installing the plugin.|

Please see the [documentation](docs/guides/extend.md) about the plugin system.

## Limitations

TFLint load configurations in the same way as Terraform v0.13. This means that it cannot inspect configurations that cannot be parsed on Terraform v0.13.

See [Compatibility with Terraform](docs/guides/compatibility.md) for details.

## Usage

TFLint inspects all configurations under the current directory by default. You can also change the behavior with the following options:

```
$ tflint --help
Usage:
  tflint [OPTIONS] [FILE or DIR...]

Application Options:
  -v, --version                                   Print TFLint version
      --langserver                                Start language server
  -f, --format=[default|json|checkstyle|junit]    Output format (default: default)
  -c, --config=FILE                               Config file name (default: .tflint.hcl)
      --ignore-module=SOURCE                      Ignore module sources
      --enable-rule=RULE_NAME                     Enable rules from the command line
      --disable-rule=RULE_NAME                    Disable rules from the command line
      --only=RULE_NAME                            Enable only this rule, disabling all other defaults. Can be specified multiple times
      --var-file=FILE                             Terraform variable file name
      --var='foo=bar'                             Set a Terraform variable
      --module                                    Inspect modules
      --deep                                      Enable deep check mode
      --aws-access-key=ACCESS_KEY                 AWS access key used in deep check mode
      --aws-secret-key=SECRET_KEY                 AWS secret key used in deep check mode
      --aws-profile=PROFILE                       AWS shared credential profile name used in deep check mode
      --aws-creds-file=FILE                       AWS shared credentials file path used in deep checking
      --aws-region=REGION                         AWS region used in deep check mode
      --force                                     Return zero exit status even if issues found
      --no-color                                  Disable colorized output
      --loglevel=[trace|debug|info|warn|error]    Change the loglevel (default: none)

Help Options:
  -h, --help                                      Show this help message
```

See [User guide](docs/guides) for each option.

## Exit Statuses

TFLint returns the following exit statuses on exit:

- 0: No issues found
- 2: Errors occurred
- 3: No errors occurred, but issues found

## FAQ
### Does TFLint check modules recursively?
- No. TFLint always checks only the current root module (no recursive check)

### Do I need to install Terraform for TFLint to work?
- No. TFLint works as a single binary because Terraform is embedded as a library. Note that this means that the version of Terraform used is determined for each TFLint version. See also [Compatibility with Terraform](docs/guides/compatibility.md). 

### TFLint causes a loading error in my code that is valid in Terraform. Why?
- First, check the version of Terraform you are using. Terraform v0.12 introduced a major syntax change, and unfortunately TFLint only supports that new syntax.

## Debugging

If you don't get the expected behavior, you can see the detailed logs when running with `TFLINT_LOG` environment variable.

```console
$ TFLINT_LOG=debug tflint
```

## Developing

See [Developer guide](docs/DEVELOPING.md).
