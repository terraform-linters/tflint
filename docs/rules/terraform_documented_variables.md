# terraform_documented_variables

Disallow `variable` declarations without description.

## Example

```hcl
variable "no_description" {
  default = "value"
}

variable "empty_description" {
  default = "value"
  description = ""
}

variable "description" {
  default = "value"
  description = "This is description"
}
```

```
$ tflint
2 issue(s) found:

Notice: `no_description` variable has no description (terraform_documented_variables)

  on template.tf line 1:
   1: variable "no_description" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_documented_variables.md

Notice: `empty_description` variable has no description (terraform_documented_variables)

  on template.tf line 5:
   5: variable "empty_description" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_documented_variables.md
 
```

## Why

Since `description` is optional value, it is not always necessary to write it. But this rule is useful if you want to force the writing of description. Especially it is useful when combined with [terraform-docs](https://github.com/segmentio/terraform-docs).

## How To Fix

Write a description other than an empty string.
