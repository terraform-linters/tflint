# TFLint
[![Build Status](https://travis-ci.org/wata727/tflint.svg?branch=master)](https://travis-ci.org/wata727/tflint)
[![GitHub release](https://img.shields.io/github/release/wata727/tflint.svg)](https://github.com/wata727/tflint/releases/latest)
[![Docker Hub](https://img.shields.io/badge/docker-ready-blue.svg)](https://hub.docker.com/r/wata727/tflint/)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/wata727/tflint)](https://goreportcard.com/report/github.com/wata727/tflint)

TFLint is a [Terraform](https://www.terraform.io/) linter focused on possible errors, best practices, and so on.

## Why TFLint is required?

Terraform is a great tool for Infrastructure as Code. However, many of these tools don't validate provider-specific issues. For example, see the following configuration file:

```hcl
resource "aws_instance" "web" {
  ami           = "ami-b73b63a0"
  instance_type = "t1.2xlarge" # invalid type!

  tags {
    Name = "HelloWorld"
  }
}
```

Since `t1.2xlarge` is a nonexistent instance type, an error will occur when you run `terraform apply`. But `terraform plan` and `terraform validate` cannot find this possible error beforehand. That's because it's an AWS provider-specific issue and it's valid as a Terraform configuration.

TFLint finds such errors in advance:

```
$ tflint
template.tf
        ERROR:3 "t1.2xlarge" is invalid instance type. (aws_instance_invalid_type)

Result: 2 issues  (1 errors , 0 warnings , 1 notices)
```

## Installation

You can download the binary built for your architecture from [the latest release](https://github.com/wata727/tflint/releases/latest). The following is an example of installation on macOS:

```
$ wget https://github.com/wata727/tflint/releases/download/v0.8.3/tflint_darwin_amd64.zip
$ unzip tflint_darwin_amd64.zip
Archive:  tflint_darwin_amd64.zip
  inflating: tflint
$ mkdir -p /usr/local/tflint/bin
$ export PATH=/usr/local/tflint/bin:$PATH
$ install tflint /usr/local/tflint/bin
$ tflint -v
```

For Linux based OS, you can use the [`install_linux.sh`](https://raw.githubusercontent.com/wata727/tflint/master/install_linux.sh) to automate the installation process.

### Homebrew

macOS users can also use [Homebrew](https://brew.sh) to install TFLint:

```
$ brew tap wata727/tflint
$ brew install tflint
```

### Docker

You can also use [TFLint via Docker](https://hub.docker.com/r/wata727/tflint/).

```
$ docker run --rm -v $(pwd):/data -t wata727/tflint
```

## Features

See [Rules](docs/rules).

## Limitations

TFLint currently only inspect Terraform-specific issues and AWS issues.

Also, load configurations in the same way as Terraform v0.12. This means that it cannot inspect configurations that cannot be parsed on Terraform v0.12.

[Named values](https://www.terraform.io/docs/configuration/expressions.html#references-to-named-values) are supported only for [input variables](https://www.terraform.io/docs/configuration/variables.html) and [workspaces](https://www.terraform.io/docs/state/workspaces.html). Expressions that contain anything else are excluded from the  inspection. [Built-in Functions](https://www.terraform.io/docs/configuration/functions.html) are fully supported.

## Usage

TFLint inspects all configurations under the current directory by default. You can also change the behavior with the following options:

```
$ tflint --help
Usage:
  tflint [OPTIONS] [FILE or DIR...]

Application Options:
  -v, --version                             Print TFLint version
  -f, --format=[default|json|checkstyle]    Output format (default: default)
  -c, --config=FILE                         Config file name (default: .tflint.hcl)
      --ignore-module=SOURCE1,SOURCE2...    Ignore module sources
      --ignore-rule=RULE1,RULE2...          Ignore rule names
      --var-file=FILE1,FILE2...             Terraform variable file names
      --var='foo=bar'                       Set a Terraform variable
      --deep                                Enable deep check mode
      --aws-access-key=ACCESS_KEY           AWS access key used in deep check mode
      --aws-secret-key=SECRET_KEY           AWS secret key used in deep check mode
      --aws-profile=PROFILE                 AWS shared credential profile name used in deep check mode
      --aws-region=REGION                   AWS region used in deep check mode
      --error-with-issues                   Return error code when issues exist
  -q, --quiet                               Do not output any message when no issues are found (default format only)

Help Options:
  -h, --help                                Show this help message
```

### Config file

By default, TFLint looks up `.tflint.hcl` according to the following priority:

- Current directory (`./.tflint.hcl`)
- Home directory (`~/.tflint.hcl`)

The config file is written in [HCL](https://github.com/hashicorp/hcl), and you can use this file instead of passing command line options.

```hcl
config {
  deep_check = true

  aws_credentials = {
    access_key = "AWS_ACCESS_KEY"
    secret_key = "AWS_SECRET_KEY"
    region     = "us-east-1"
  }

  ignore_module = {
    "github.com/wata727/example-module" = true
  }

  varfile = ["example1.tfvars", "example2.tfvars"]

  variables = ["foo=bar", "bar=[\"baz\"]"]
}

rule "aws_instance_invalid_type" {
  enabled = false
}

rule "aws_instance_previous_type" {
  enabled = false
}
```

You can also use another file as a config file with the `--config` option.

```
$ tflint --config other_config.hcl
```

### Rules

You can make settings for each rule in the `rule` block. Currently, it can set only `enabled` option. If you set `enabled = false`, TFLint doesn't inspect configuration files by this rule.

```hcl
rule "aws_instance_previous_type" {
  enabled = false
}
```

You can also disable rules with the `--ignore-rule` option.

```
$ tflint --ignore-rule=aws_instance_invalid_type,aws_instance_previous_type
```

Also, annotation comments can disable rules on specific lines:

```hcl
resource "aws_instance" "foo" {
    # tflint-ignore: aws_instance_invalid_type
    instance_type = "t1.2xlarge"
}
```

The annotation works only for the same line or the line below it. You can also use `tflint-ignore: all` if you want to ignore all the rules.

See also [list of available rules](docs/rules).

### Deep Checking

When deep checking is enabled, TFLint invokes the provider's API to do a more detailed inspection. For example, find a non-existent IAM profile name etc. You can enable it with the `--deep` option.

```
$ tflint --deep
template.tf
        ERROR:3 "t1.2xlarge" is invalid instance type. (aws_instance_invalid_type)
        ERROR:4 "invalid_profile" is invalid IAM profile name. (aws_instance_invalid_iam_profile)

Result: 2 issues  (2 errors , 0 warnings , 0 notices)
```

In order to enable deep checking, [credentials](#credentials) are needed.

### Credentials

TFLint supports various credential providers. It is used with the following priority:

- Static credentials
- Shared credentials
- Environment credentials
- Default shared credentials

#### Static Credentials

If you have an access key and a secret key, you can pass these keys.

```
$ tflint --aws-access-key AWS_ACCESS_KEY --aws-secret-key AWS_SECRET_KEY --aws-region us-east-1
```

```hcl
config {
  aws_credentials = {
    access_key = "AWS_ACCESS_KEY"
    secret_key = "AWS_SECRET_KEY"
    region     = "us-east-1"
  }
}
```

#### Shared Credentials

If you have [shared credentials](https://aws.amazon.com/jp/blogs/security/a-new-and-standardized-way-to-manage-credentials-in-the-aws-sdks/), you can pass the profile name. However, only `~/.aws/credentials` is supported as a credential location.

```
$ tflint --aws-profile AWS_PROFILE --aws-region us-east-1
```

```hcl
config {
  aws_credentials = {
    profile = "AWS_PROFILE"
    region  = "us-east-1"
  }
}
```

#### Environment Credentials

TFLint looks up `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` environment variables. This is useful when you don't want to explicitly pass credentials.

```
$ export AWS_ACCESS_KEY_ID=AWS_ACCESS_KEY
$ export AWS_SECRET_ACCESS_KEY=AWS_SECRET_KEY
```

### Module Inspection

TFLint can also inspect [modules](https://www.terraform.io/docs/configuration/modules.html). In this case, it checks based on the input variables passed to the calling module.

```hcl
module "aws_instance" {
  source        = "./module"

  ami           = "ami-b73b63a0"
  instance_type = "t1.2xlarge"
}
```

```
$ tflint
aws_instance/main.tf
        ERROR:6 "t1.2xlarge" is invalid instance type. (aws_instance_invalid_type)

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

TFLint loads modules in the same way as Terraform. So note that you need to run `terraform init` first.

You can use the `--ignore-module` option if you want to skip inspection for a particular module. Note that you need to pass module sources rather than module ids for backward compatibility.

```
$ tflint --ignore-module=./module
```

### Run with a specific configuration file

If you want to inspect only a specific configuration file, not all files, you can pass a file as an argument.

```
$ tflint main.tf
```

## Debugging

If you don't get the expected behavior, you can see the detailed logs when running with `TFLINT_LOG` environment variable.

```
$ TFLINT_LOG=debug tflint
```

## Developing

See [Developer Guides](docs/DEVELOPING.md).

## Author

[Kazuma Watanabe](https://github.com/wata727)
