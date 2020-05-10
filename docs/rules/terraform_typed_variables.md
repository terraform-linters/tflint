# terraform_typed_variables
Disallow `variable` declarations without type.

**NOTE**: Due to the current implementation of TFLint, it is not able to distinguish between the case where
no type is declared and `type = any` as Terraform treats these two cases the same. If this rule is enabled
it will report a warning for variables declared with `type = any`.
```
variable "any_type" {
    description = "A variable with 'any' type declared will generate a warning"
    type        = any
}
```

A future version of TFLint may look to implement lower level checks of the HCL syntax. 
See https://github.com/terraform-linters/tflint/issues/741 to track this issue.


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
