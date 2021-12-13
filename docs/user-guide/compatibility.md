# Compatibility with Terraform

TFLint bundles Terraform internal packages as a library. This allows the Terraform language to be parsed correctly even if Terraform is not installed at runtime.

On the other hand, language semantics depend on the behavior of a particular bundled version. For example, a configuration parsed by Terraform v1.0 may be parsed by v1.1 language parser. The currently bundled version is v1.1.0.

The best practice is to match the Terraform version bundled with TFLint to the version you actually use. However, the Terraform language guarantees some backward compatibility, so different versions may not cause immediate problems. However, keep in mind that false positives/negatives can occur depending on this assumption.

## Input Variables

Like Terraform, it supports the `--var`,` --var-file` options, environment variables (`TF_VAR_*`), and automatically loading variable definitions (`terraform.tfvars` and `*.auto.tfvars`) files. See [Input Variables](https://www.terraform.io/docs/language/values/variables.html).

Input variables are evaluated correctly, just like Terraform:

```hcl
variable "instance_type" {
  default = "t2.micro"
}

resource "aws_instance" "foo" {
  instance_type = var.instance_type # => "t2.micro"
}
```

## Named Values

[Named values](https://www.terraform.io/docs/configuration/expressions/references.html) are supported partially. The following named values are available:

- `var.<NAME>`
- `path.module`
- `path.root`
- `path.cwd`
- `terraform.workspace`

Expressions that reference named values not included above (e.g. `locals.*`, `count.*`, `each.*`, etc.) are excluded from the inspection.

```hcl
locals {
  instance_family = "t2"
}

resource "aws_instance" "foo" {
  instance_type = "${local.instance_family}.micro" # => Not an error, it will be ignored because it marks as unknown
}
```

## Built-in Functions

[Built-in Functions](https://www.terraform.io/docs/configuration/functions.html) are fully supported.

## Environment Variables

The following environment variables are supported:

- [TF_VAR_name](https://www.terraform.io/docs/commands/environment-variables.html#tf_var_name)
- [TF_DATA_DIR](https://www.terraform.io/docs/commands/environment-variables.html#tf_data_dir)
- [TF_WORKSPACE](https://www.terraform.io/docs/commands/environment-variables.html#tf_workspace)
