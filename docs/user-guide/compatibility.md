# Compatibility with Terraform

TFLint interprets the [Terraform language](https://developer.hashicorp.com/terraform/language) with its own parser which is a fork of the Terraform's native one. This allows it to be parsed correctly even if Terraform is not installed at runtime.

The parser supports Terraform v1.x syntax and semantics. The language compatibility on Terraform v1.x is defined by [Compatibility Promises](https://developer.hashicorp.com/terraform/language/v1-compatibility-promises). TFLint follows this promise. New features are only supported in newer TFLint versions, and bug and experimental features compatibility are not guaranteed.

The latest supported version is Terraform v1.9.

## Input Variables

Like Terraform, TFLint supports the `--var`,` --var-file` options, environment variables (`TF_VAR_*`), and automatically loading variable definitions (`terraform.tfvars` and `*.auto.tfvars`) files. See [Input Variables](https://developer.hashicorp.com/terraform/language/values/variables).

Input variables are evaluated just like in Terraform:

```hcl
variable "instance_type" {
  default = "t2.micro"
}

resource "aws_instance" "foo" {
  instance_type = var.instance_type # => "t2.micro"
}
```

Unknown variables (e.g. no default) are ignored:

```hcl
variable "instance_type" {}

resource "aws_instance" "foo" {
  instance_type = var.instance_type # => ignored
}
```

Sensitive variables are ignored. This is to avoid unintended disclosure.

```hcl
variable "instance_type" {
  sensitive = true
  default   = "t2.micro"
}

resource "aws_instance" "foo" {
  instance_type = var.instance_type # => ignored
}
```

## Local Values

TFLint supports [Local Values](https://developer.hashicorp.com/terraform/language/values/locals).

```hcl
variable "foo" {
  default = "variable value"
}

locals {
  static   = "static value"
  variable = var.foo
  local    = local.static
  resource = aws_instance.main.arn
}

local.static   # => "static value"
local.variable # => "variable value"
local.local    # => "static value"
local.resource # => ignored (unknown)
```

## The `count` and `for_each` Meta-Arguments

TFLint supports the [`count`](https://developer.hashicorp.com/terraform/language/meta-arguments/count) and [`for_each`](https://developer.hashicorp.com/terraform/language/meta-arguments/for_each) meta-arguments.

```hcl
resource "aws_instance" "foo" {
  count = 0

  instance_type = "invalid" # => ignored because ths resource is not created
}
```

```hcl
resource "aws_instance" "foo" {
  count = 2

  instance_type = "t${count.index}.micro" # => "t0.micro" and "t1.micro"
}
```

Note that this behavior may differ depending on a rule. Rules like `terraform_deprecated_syntax` will check resources regardless of the meta-argument values.

If the meta-arguments are unknown, the resource/module is ignored:

```hcl
variable "count" {}

resource "aws_instance" "foo" {
  count = var.count

  instance_type = "invalid" # => ignored
}
```

## The `path.*` and `terraform.workspace` Values

TFLint supports [filesystem and workspace info](https://developer.hashicorp.com/terraform/language/expressions/references#filesystem-and-workspace-info).

- `path.module`
- `path.root`
- `path.cwd`
- `terraform.workspace`.

## Unsupported Named Values

The values below are state-dependent and cannot be determined statically, so TFLint resolves them to unknown values.

- `<RESOURCE TYPE>.<NAME>`
- `module.<MODULE NAME>`
- `data.<DATA TYPE>.<NAME>`
- `self`

## Functions

[Built-in Functions](https://developer.hashicorp.com/terraform/language/functions) are fully supported. However, functions such as [`plantimestamp`](https://developer.hashicorp.com/terraform/language/functions/plantimestamp) whose return value cannot be determined statically will return an unknown value.

[Provider-defined functions](https://www.hashicorp.com/blog/terraform-1-8-adds-provider-functions-for-aws-google-cloud-and-kubernetes) always return unknown values, except for `provider::terraform::*` functions.

## Dynamic Blocks

TFLint supports [dynamic blocks](https://developer.hashicorp.com/terraform/language/expressions/dynamic-blocks).

```hcl
resource "aws_instance" "dynamic" {
  dynamic "ebs_block_device" {
    for_each = toset([
      { size = 10 },
      { size = 20 }
    ])
    content {
      volume_size = ebs_block_device.value["size"] # => 10 and 20
    }
  }
}
```

Similar to support for meta-arguments, some rules may process a dynamic block as-is without expansion. If the `for_each` is unknown, the block will be empty.

## Modules

TFLint doesn't automatically inspect the content of modules themselves. However, by default, it will analyze their content in order to raise any issues that arise from attributes in module calls.

```hcl
resource "aws_instance" "static" {
  ebs_block_device {
    encrypted = false # => Must be encrypted
  }
}

module "aws_instance" {
  source = "./module/aws_instance"

  encrypted = false # => Must be encrypted
}
```

Remote modules can also be inspected. See [Calling Modules](./calling-modules.md) for details.

## Environment Variables

The following environment variables are supported:

- [TF_VAR_name](https://developer.hashicorp.com/terraform/cli/config/environment-variables#tf_var_name)
- [TF_DATA_DIR](https://developer.hashicorp.com/terraform/cli/config/environment-variables#tf_data_dir)
- [TF_WORKSPACE](https://developer.hashicorp.com/terraform/cli/config/environment-variables#tf_workspace)
