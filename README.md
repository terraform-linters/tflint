# TFLint
[![Build Status](https://travis-ci.org/wata727/tflint.svg?branch=master)](https://travis-ci.org/wata727/tflint)
[![GitHub release](https://img.shields.io/github/release/wata727/tflint.svg)](https://github.com/wata727/tflint/releases/latest)
[![Docker Hub](https://img.shields.io/badge/docker-ready-blue.svg)](https://hub.docker.com/r/wata727/tflint/)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

TFLint is [Terraform](https://www.terraform.io/) linter for detecting errors that can not be detected by `terraform plan`

## Why TFLint is Required?
Terraform is a great tool for infrastructure as a code. It generates an execution plan, we can rely on this plan to proceed with development. However, this plan does not verify values used in template. For example, following template is invalid configuration (t1.2xlarge is invalid instance type)

```
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
Download binary built for your architecture from [latest releases](https://github.com/wata727/tflint/releases/latest). After downloading, place the binary on the directory on the PATH. An example of installation by command is as follows.
```
$ wget https://github.com/wata727/tflint/releases/download/v0.3.1/tflint_darwin_amd64.zip
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
$ brew tap hakamadare/tflint
$ brew install tflint
```

### Running in Docker
We provide Docker images for each version on [DockerHub](https://hub.docker.com/r/wata727/tflint/). With docker, you can run TFLint without installing it locally.

```
$ docker run -v $(pwd):/data --workdir=/data -t wata727/tflint
```

## Quick Start
Try running TFLint under the directory where Terraform is executed. It detect if there is a issue and output the result. For example, run on the previous invalid template.

```
$ tflint
template.tf
        NOTICE:1 "iam_instance_profile" is not specified. If you want to change it, you need to recreate it
        ERROR:3 "t1.2xlarge" is invalid instance type.

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
--var-file <file1,file2...>             specify terraform variable files.
--deep                                  enable deep check mode.
--aws-access-key                        set AWS access key used in deep check mode.
--aws-secret-key                        set AWS secret key used in deep check mode.
--aws-region                            set AWS region used in deep check mode.
-d, --debug                             enable debug mode.
--error-with-issues                     return error code when issue exists.
--fast                                  ignore slow rules. currently, ignore only 'aws_instance_invalid_ami'
```

## Configuration
By default, TFLint loads `.tflint.hcl` under the current directory. The configuration file is described in [HCL](https://github.com/hashicorp/hcl), and options available on the command line can be described in advance. Following example:

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

  varfile = ["example1.tfvars", "example2.tfvars"]
}
```

If you want to create a configuration file with a different name, specify the file name with `--config` option.

```
$ tflint --config other_config.hcl
```

## Interpolation Syntax Support
TFLint can interpret part of [interpolation syntax](https://www.terraform.io/docs/configuration/interpolation.html). We now support only variables. So you cannot use attributes of resource, outputs of modules and built-in functions. If you are using them, TFLint ignores it. You can check what is ignored by executing it with `--debug` option.

### Variable Files
If you use [variable files](https://www.terraform.io/docs/configuration/variables.html#variable-files), Please specify it by arguments or configuration file. TFLint interprets variables as well as Terraform. In other words, when variables are conflicting, It will be overridden or merged correctly.

## Deep Check
Deep check is an option that you can actually search resources on AWS and check invalid references and duplicate resources. You can activate it by executing it with `--deep` option as following:

```
$ tflint --deep
template.tf
        ERROR:3 "t1.2xlarge" is invalid instance type.
        ERROR:4 "invalid_profile" is invalid IAM profile name.

Result: 2 issues  (2 errors , 0 warnings , 0 notices)
```

In the above example, an IAM instance profile that does not actually exist is specified, so it is an error. In order to refer to actual resources, AWS credentials are required. You can use command line options, configuration files, environment variables, shared credentials for these specifications.

## Developing
If you want to build TFLint at your environment, you can build with the following procedure. [Go](https://golang.org/) 1.8 or more is required.

```
$ make build
go get github.com/Masterminds/glide
glide install
[INFO]  Downloading dependencies. Please wait...
...
go test $(go list ./... | grep -v vendor | grep -v mock)
...
go build -v
...
github.com/wata727/tflint
```

## Author

[Kazuma Watanabe](https://github.com/wata727)
