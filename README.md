# TFLint
[![Build Status](https://travis-ci.org/wata727/tflint.svg?branch=master)](https://travis-ci.org/wata727/tflint)
[![GitHub release](https://img.shields.io/github/release/wata727/tflint.svg)](https://github.com/wata727/tflint/releases/latest)
[![Docker Hub](https://img.shields.io/badge/docker-ready-blue.svg)](https://hub.docker.com/r/wata727/tflint/)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

TFLint is a [Terraform](https://www.terraform.io/) linter for detecting errors that can not be detected by `terraform plan`

## Current Project Status

Currently, we are not encouraged to run TFLint on critical workflow due to incomplete features. See [#167](https://github.com/wata727/tflint/issues/167), [#168](https://github.com/wata727/tflint/issues/168).  
This issue is scheduled to be fixed in the release of v0.8.0, but this release depends on Terraform v0.12.0, and we cannot release v0.8.0 until it is released.  
For these reasons, feature requests and major changes have been suspended until v0.8.0 release. A small bug fixes for patch upgrade are welcome.

## Why TFLint is Required?
Terraform is a great tool for infrastructure as a code. It generates an execution plan, we can rely on this plan to proceed with development. However, this plan does not verify values used in template. For example, following template is invalid configuration (t1.2xlarge is invalid instance type)

```hcl
resource "aws_instance" "web" {
  ami           = "ami-b73b63a0"
  instance_type = "t1.2xlarge" # invalid type!

  tags {
    Name = "HelloWorld"
  }
}
```

If you run `terraform apply` for this template, it will obviously produce an error. However, `terraform plan` can get an execution plan without causing an error. This is often not a desirable result. In order to solve this problem, TFLint validates values used in template.

## Installation
Download binary built for your architecture from [latest releases](https://github.com/wata727/tflint/releases/latest). After downloading, place the binary on the directory on the PATH. The following example is the installation in macOS.
```
$ wget https://github.com/wata727/tflint/releases/download/v0.7.4/tflint_darwin_amd64.zip
$ unzip tflint_darwin_amd64.zip
Archive:  tflint_darwin_amd64.zip
  inflating: tflint
$ mkdir -p /usr/local/tflint/bin
$ export PATH=/usr/local/tflint/bin:$PATH
$ install tflint /usr/local/tflint/bin
$ tflint -v
```

### Homebrew

macOS users can also use [Homebrew](https://brew.sh) to install TFLint:

```
$ brew tap wata727/tflint
$ brew install tflint
```

### Running in Docker
We provide Docker images for each version on [DockerHub](https://hub.docker.com/r/wata727/tflint/). With docker, you can run TFLint without installing it locally.

```
$ docker run -itv (pwd):/project -w /project --rm wata727/tflint tflint
```

## Quick Start
Try running TFLint under the directory where Terraform is executed. It detect if there is a issue and output the result. For example, run on the previous invalid template.

```
$ tflint
template.tf
        ERROR:3 "t1.2xlarge" is invalid instance type. (aws_instance_invalid_type)

Result: 2 issues  (1 errors , 0 warnings , 1 notices)
```

If you would like to know more about these issues and available features please check the [documentation](https://github.com/wata727/tflint/tree/master/docs).

### Specify Template
If you want to parse only a specific template, not all templates, you can specify a filename as an argument.

```
$ tflint template.tf
```

## Available Options
Please show `tflint --help`

```
Usage:
  tflint [OPTIONS] [FILE]

Application Options:
  -v, --version                             Print TFLint version
  -f, --format=[default|json|checkstyle]    Output format (default: default)
  -c, --config=FILE                         Config file name (default: .tflint.hcl)
      --ignore-module=SOURCE1,SOURCE2...    Ignore module sources
      --ignore-rule=RULE1,RULE2...          Ignore rule names
      --var-file=FILE1,FILE2...             Terraform variable file names
      --deep                                Enable deep check mode
      --aws-access-key=ACCESS_KEY           AWS access key used in deep check mode
      --aws-secret-key=SECRET_KEY           AWS secret key used in deep check mode
      --aws-profile=PROFILE                 AWS shared credential profile name used in deep check mode
      --aws-region=REGION                   AWS region used in deep check mode
  -d, --debug                               Enable debug mode
      --error-with-issues                   Return error code when issues exist
      --fast                                Ignore slow rules. Currently, ignore only aws_instance_invalid_ami
  -q, --quiet                               Do not output any message when no issues are found (Format=default only)

Help Options:
  -h, --help                                Show this help message
```

## Configuration
By default, TFLint loads `.tflint.hcl` according to the following priority:

- Current directory (`./.tflint.hcl`)
- Home directory (`~/.tflint.hcl`)

The configuration file is described in [HCL](https://github.com/hashicorp/hcl), and options available on the command line can be described in advance. Following example:

```hcl
config {
  terraform_version = "0.9.11"
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
}

rule "aws_instance_invalid_type" {
  enabled = false
}

rule "aws_instance_previous_type" {
  enabled = false
}
```

If you want to create a configuration file with a different name, specify the file name with `--config` option.

```
$ tflint --config other_config.hcl
```

### Terraform Version
You can set the version of Terraform you are using. If it is set, TFLint will detect issues according to it.

### Credentials
TFLint supports various credential providers. It is used with the following priority:

- Static credentials
- Shared credentials
- Environment credentials
- Default shared credentials

#### Static Credentials
If you have access key and secret key, you can specify these credentials.

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
If you have [shared credentials](https://aws.amazon.com/jp/blogs/security/a-new-and-standardized-way-to-manage-credentials-in-the-aws-sdks/), you can specify credentials profile name. However TFLint supports only `~/.aws/credentials` as shared credentials location.

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
TFLint looks `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`, `AWS_REGION` environment values. This is useful when you do not want to explicitly specify credentials.

```
$ export AWS_ACCESS_KEY_ID=AWS_ACCESS_KEY
$ export AWS_SECRET_ACCESS_KEY=AWS_SECRET_KEY
```

### Rules

You can make settings for each rule in the `rule` block. Currently, it can set only `enabled` option. If you set `enabled = false`, TFLint doesn't check templates by this rule.

```
rule "aws_instance_previous_type" {
  enabled = false
}
```

Please see the [documentation](https://github.com/wata727/tflint/tree/master/docs) for a list of rules.

## Interpolation Syntax Support
TFLint can interpret part of [interpolation syntax](https://www.terraform.io/docs/configuration/interpolation.html). We now support only variables and terraform meta information (e.g. "${terraform.env}"). So you cannot use attributes of resource, outputs of modules and built-in functions. If you are using them, TFLint ignores it. You can check what is ignored by executing it with `--debug` option.

### Variable Files
If you use [variable files](https://www.terraform.io/docs/configuration/variables.html#variable-files), Please specify it by arguments or configuration file. TFLint interprets variables as well as Terraform. In other words, when variables are conflicting, It will be overridden or merged correctly.

## Deep Check
Deep check is an option that you can actually search resources on AWS and check invalid references and duplicate resources. You can activate it by executing it with `--deep` option as following:

```
$ tflint --deep
template.tf
        ERROR:3 "t1.2xlarge" is invalid instance type. (aws_instance_invalid_type)
        ERROR:4 "invalid_profile" is invalid IAM profile name. (aws_instance_invalid_iam_profile)

Result: 2 issues  (2 errors , 0 warnings , 0 notices)
```

In the above example, an IAM instance profile that does not actually exist is specified, so it is an error. In order to refer to actual resources, AWS credentials are required. You can use command line options, configuration files, environment variables, shared credentials for these specifications.

## Developing
If you want to build TFLint at your environment, you can build with the following procedure. [Go](https://golang.org/) 1.9 or more is required.

```
$ make build
```

## Author

[Kazuma Watanabe](https://github.com/wata727)
