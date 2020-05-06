# terraform_unused_declarations

Disallow variables, data sources, and locals that are declared but never used.

## Example

```hcl
variable "not_used" {}

variable "used" {}
output "out" {
  value = var.used
}
```

```
$ tflint
1 issue(s) found:

Warning: variable "not_used" is declared but not used (terraform_unused_declarations)

  on config.tf line 1:
   1: variable "not_used" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.5/docs/rules/terraform_unused_declarations.md
 
```

## Why

Terraform will ignore variables and locals that are not used. It will refresh declared data sources regardless of usage. However, unreferenced variables likely indicate either a bug (and should be referenced) or removed code (and should be removed).

## How To Fix

Remove the declaration. For `variable` and `data`, remove the entire block. For a `local` value, remove the attribute from the `locals` block.

While data sources should generally not have side effects, take greater care when removing them. For example, removing `data "http"` will cause Terraform to no longer perform an HTTP `GET` request during each plan. If a data source is being used for side effects, add an annotation to ignore it:

```tf
# tflint-ignore: terraform_unused_declarations
data "http" "example" {
  url = "https://checkpoint-api.hashicorp.com/v1/check/terraform"
}
```