# terraform_typed_variables

Disallow `variable` declarations without type.

## Example

```hcl
variable "no_type" {
  default = "value"
}

variable "enabled" {
  default     = false
  description = "This is description"
  type        = bool
}
```

```
$ tflint
1 issue(s) found:

Warning: `no_type` variable has no type (terraform_typed_variables)

  on template.tf line 1:
   1: variable "no_type" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_typed_variables.md
 
```

## Why

Since `type` is optional value, it is not always necessary to declare it. But this rule is useful if you want to force declaration of a type.

## How To Fix
Add a type to the variable. See https://www.terraform.io/docs/configuration/variables.html#type-constraints for more details about types
