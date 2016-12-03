# TFLint
[![Build Status](https://travis-ci.org/wata727/tflint.svg?branch=master)](https://travis-ci.org/wata727/tflint)
[![GitHub release](https://img.shields.io/github/release/wata727/tflint.svg)](https://github.com/wata727/tflint/releases/latest)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

TFLint is a linter of [Terraform](https://www.terraform.io/). TFLint is intended to supplement `terraform plan` in AWS provider. In particular, it can detect errors that can not be detected by `terraform plan` or template that do not follow best practices.

## Why do we need to supplement `terraform plan`?
Terraform is a great tool for infrastructure as a code. it generates an execution plan, we can rely on this plan to proceed with development. However, this plan does not verify values used in template. For example, following template is invalid configuration (t2.2xlarge is not exists)

```
resource "aws_instance" "web" {
  ami           = "ami-b73b63a0"
  instance_type = "t2.2xlarge" # invalid type!

  tags {
    Name = "HelloWorld"
  }
}
```

If you run `terraform apply` for this template, it will obviously produce an error. However, `terraform plan` can get an execution plan without causing an error. This is often not a desirable result. In order to solve this problem, TFLint validates values used in template.

## Installation
Download binary built for your architecture from [latest releases](https://github.com/wata727/tflint/releases/latest). After downloading, place the binary on the directory on the PATH.

An example of installation by command is as follows.
```
$ wget https://github.com/wata727/tflint/releases/download/v0.1.0/tflint_darwin_amd64.zip
$ unzip tflint_darwin_amd64.zip
Archive:  tflint_darwin_amd64.zip
  inflating: tflint
$ mkdir -p /usr/local/tflint/bin
$ export PATH=/usr/local/tflint/bin:$PATH
$ install tflint /usr/local/tflint/bin
$ tflint -v
```

## Quick Start
Try running TFLint under the directory where Terraform is executed. it detect if there is a issue and output the result. For example, we analyze the previous invaild template.

```
$ tflint
template.tf
        NOTICE:1 "iam_instance_profile" is not specified. If you want to change it, you need to recreate it
        ERROR:3 "t2.2xlarge" is invalid instance type.

Result: 2 issues  (1 errors , 0 warnings , 1 notices)
```

Two issues were reported. One is for invalid instance type, the other one is for beat practices about IAM instance profile. If you would like to know more about these issues please check the [documentation](https://github.com/wata727/tflint/tree/master/docs).

### Specify Template
If you want to parse only a specific template, not all templates, you can specify a filename as an argument.

```
$ tflint template.tf
```

## Available Options
Please show `tflint --help`

```
-h, --help                              show usage of TFLint. This page.
-v, --version                           print version information.
-f, --format <format>                   choose output format from "default" or "json"
-c, --config <file>                     specify config file. default is ".tflint.hcl"
--ignore-module <source1,source2...>    ignore module by specified source.
--ignore-rule <rule1,rule2...>          ignore rules.
--deep                                  enable deep check mode.
--aws-access-key                        set AWS access key used in deep check mode.
--aws-secret-key                        set AWS secret key used in deep check mode.
--aws-region                            set AWS region used in deep check mode.
-d, --debug                             enable debug mode.
```

## Configuration
By default, TFLint reads ".tflint.hcl" under the current directory. The configuration file is described in [HCL](https://github.com/hashicorp/hcl), and options available on the command line can be described in advance. Following example:

```
config {
  deep_check = true

  aws_credentials = {
    access_key = "AWS_ACCESS_KEY"
    secret_key = "AWS_SECRET_KEY"
    region     = "us-east-1"
  }

  ignore_rule = {
    aws_instance_invalid_type  = true
    aws_instance_previous_type = true
  }

  ignore_module = {
    "github.com/wata727/example-module" = true
  }
}
```

If you want to create a configuration file with a different name, specify the file name with `--config` option.

```
$ tflint --config other_config.hcl
```

## Interpolation Syntax Support
TFLint can interpret part of [interpolation syntax](https://www.terraform.io/docs/configuration/interpolation.html). We now support only variables. So you cannot use attributes of resource, outputs of modules and built-in functions. If you are using them, TFLint ignores it. You can check what is ignored by executing it with `--debug` option.

## Deep Check?
Deep check is an option that you can actually search resources on AWS and check if invalid values are used. You can activate it by executing it with `--deep` option as following:

```
$ tflint --deep
template.tf
        ERROR:3 "t2.2xlarge" is invalid instance type.
        ERROR:4 "invalid_profile" is invalid IAM profile name.

Result: 2 issues  (2 errors , 0 warnings , 0 notices)
```

In the above example, an IAM instance profile that does not actually exist is specified, so it is an error. In order to refer to actual resources, AWS credentials are required. You can use arguments, configuration files, environment variables, shared credentials for these specifications.

## Developing in your machine
If you want to build TFLint at your machine, you can build with the following procedure. [Go](https://golang.org/) 1.7 or more is required.

```
$ make build
go get github.com/Masterminds/glide
glide install
[INFO]  Downloading dependencies. Please wait...
...
go test $(go list ./... | grep -v vendor | grep -v mock)
ok      github.com/wata727/tflint       0.064s
ok      github.com/wata727/tflint/config        0.363s
ok      github.com/wata727/tflint/detector      0.085s
ok      github.com/wata727/tflint/evaluator     0.299s
ok      github.com/wata727/tflint/issue 0.266s
ok      github.com/wata727/tflint/loader        0.179s
ok      github.com/wata727/tflint/logger        0.086s
ok      github.com/wata727/tflint/printer       0.085s
go build -v
...
github.com/wata727/tflint
```

## Author

[Kazuma Watanabe](https://github.com/wata727)
