# terraform_standard_module_structure

Ensure that a module complies with the Terraform [Standard Module Structure](https://www.terraform.io/docs/modules/index.html#standard-module-structure)

## Example

_main.tf_
```hcl
variable "v" {}
```

```
$ tflint
1 issue(s) found:

Warning: variable "v" should be moved from main.tf to variables.tf (terraform_standard_module_structure)

  on main.tf line 1:
   1: variable "v" {}

Reference: https://github.com/terraform-linters/tflint/blob/v0.16.0/docs/rules/terraform_standard_module_structure.md
```

## Why

Terraform's documentation outlines a [Standard Module Structure](https://www.terraform.io/docs/modules/structure.html). A minimal module should have a `main.tf`, `variables.tf`, and `outputs.tf` file. Variable and output blocks should be included in the corresponding file.

## How To Fix

* Move blocks to their conventional files as needed
* Create empty files even if no `variable` or `output` blocks are defined
