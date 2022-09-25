# Compatibility with Terraform

TFLint interprets the [Terraform language](https://www.terraform.io/language) with its own parser which is a fork of the Terraform's native one. This allows it to be parsed correctly even if Terraform is not installed at runtime.

The parser supports Terraform v1.x syntax and semantics. The language compatibility on Terraform v1.x is defined by [Compatibility Promises](https://www.terraform.io/language/v1-compatibility-promises). TFLint follows this promise. New features are only supported in newer TFLint versions, and bug and experimental features compatibility are not guaranteed.

## Input Variables

Like Terraform, TFLint supports the `--var`,` --var-file` options, environment variables (`TF_VAR_*`), and automatically loading variable definitions (`terraform.tfvars` and `*.auto.tfvars`) files. See [Input Variables](https://www.terraform.io/language/values/variables).

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

## Named Values

[Named values](https://www.terraform.io/language/expressions/references) are supported partially. The following named values are available:

- `var.<NAME>`
- `path.module`
- `path.root`
- `path.cwd`
- `terraform.workspace`

Unsupported named values (e.g. `local.*`, `count.index`, `each.key`) are ignored:

```hcl
locals {
  instance_family = "t2"
}

resource "aws_instance" "foo" {
  instance_type = "${local.instance_family}.micro" # => ignored
}
```

## Built-in Functions

[Built-in Functions](https://www.terraform.io/language/functions) are fully supported.

## Conditional Resources/Modules

Resources and modules with [`count = 0`](https://www.terraform.io/language/meta-arguments/count) or [`for_each = {}`](https://www.terraform.io/language/meta-arguments/for_each) are ignored:

```hcl
resource "aws_instance" "foo" {
  count = 0

  instance_type = "invalid" # => ignored
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

## Dynamic Blocks

[Dynamic blocks](https://www.terraform.io/language/expressions/dynamic-blocks
) work just like normal blocks:

```hcl
resource "aws_instance" "static" {
  ebs_block_device {
    encrypted = false # => Must be encrypted
  }
}

resource "aws_instance" "dynamic" {
  dynamic "ebs_block_device" {
    for_each = var.block_devices
    content {
      encrypted = false # => Must be encrypted
    }
  }
}
```

Note that iterator evaluation is not supported.

```hcl
resource "aws_instance" "dynamic" {
  dynamic "ebs_block_device" {
    for_each = var.block_devices
    content {
      encrypted = ebs_block_device.value["encrypted"] # => ignored
    }
  }
}
```

## Modules

Resources contained within modules are ignored by default, but when the [Module Inspection](./module-inspection.md) is enabled, the arguments of module calls are inspected.

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

## Environment Variables

The following environment variables are supported:

- [TF_VAR_name](https://www.terraform.io/cli/config/environment-variables#tf_var_name)
- [TF_DATA_DIR](https://www.terraform.io/cli/config/environment-variables#tf_data_dir)
- [TF_WORKSPACE](https://www.terraform.io/cli/config/environment-variables#tf_workspace)
