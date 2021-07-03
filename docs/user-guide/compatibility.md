# Compatibility with Terraform

Since TFLint embeds a specific version of Terraform as a library, some features implicitly assume the behavior of Terraform v1.0.0

Of course, TFLint may work correctly if you run it on other versions. But, false positives/negatives can occur based on this assumption.

## Input Variables

Like Terraform, it supports the `--var`,` --var-file` options, automatic loading of variable definitions (`.tfvars`) files, and environment variables.

## Named Values

[Named values](https://www.terraform.io/docs/configuration/expressions/references.html) are supported partially. The following named values are available:

- `var.<NAME>`
- `path.module`
- `path.root`
- `path.cwd`
- `terraform.workspace`

Expressions that reference named values not included above are excluded from the inspection.

## Built-in Functions

[Built-in Functions](https://www.terraform.io/docs/configuration/functions.html) are fully supported.

## Environment Variables

The following environment variables are supported:

- [TF_VAR_name](https://www.terraform.io/docs/commands/environment-variables.html#tf_var_name)
- [TF_DATA_DIR](https://www.terraform.io/docs/commands/environment-variables.html#tf_data_dir)
- [TF_WORKSPACE](https://www.terraform.io/docs/commands/environment-variables.html#tf_workspace)
