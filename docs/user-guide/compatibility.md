# Compatibility with Terraform

TFLint interprets the [Terraform language](https://www.terraform.io/language) with its own parser which is a fork of the Terraform's native one. This allows it to be parsed correctly even if Terraform is not installed at runtime.

The parser supports Terraform v1.x syntax and semantics. The language compatibility on Terraform v1.x is defined by [Compatibility Promises](https://www.terraform.io/language/v1-compatibility-promises). TFLint follows this promise. New features are only supported in newer TFLint versions, and bug and experimental features compatibility are not guaranteed.

## Input Variables

Like Terraform, TFLint supports the `--var`,` --var-file` options, environment variables (`TF_VAR_*`), and automatically loading variable definitions (`terraform.tfvars` and `*.auto.tfvars`) files. See [Input Variables](https://www.terraform.io/language/values/variables).

Input variables are evaluated correctly, just like Terraform:

```hcl
variable "instance_type" {
  default = "t2.micro"
}

resource "aws_instance" "foo" {
  instance_type = var.instance_type # => "t2.micro"
}
```

Sensitive variables are ignored without being evaluated. This is to avoid unintended disclosure.

## Named Values

[Named values](https://www.terraform.io/language/expressions/references) are supported partially. The following named values are available:

- `var.<NAME>`
- `path.module`
- `path.root`
- `path.cwd`
- `terraform.workspace`

Expressions containing unsupported named values (e.g. `local.*`, `count.index`, `each.key`) are simply ignored when evaluated.

```hcl
locals {
  instance_family = "t2"
}

resource "aws_instance" "foo" {
  instance_type = "${local.instance_family}.micro" # => Not an error, it will be ignored because it marks as unknown
}
```

## Built-in Functions

[Built-in Functions](https://www.terraform.io/language/functions) are fully supported.

## Environment Variables

The following environment variables are supported:

- [TF_VAR_name](https://www.terraform.io/cli/config/environment-variables#tf_var_name)
- [TF_DATA_DIR](https://www.terraform.io/cli/config/environment-variables#tf_data_dir)
- [TF_WORKSPACE](https://www.terraform.io/cli/config/environment-variables#tf_workspace)
