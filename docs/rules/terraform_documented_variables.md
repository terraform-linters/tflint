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
variables.tf
        NOTICE:1 `no_description` variable has no description (terraform_documented_variables)
        NOTICE:5 `empty_description` variable has no description (terraform_documented_variables)
```

## Why

Since `description` is optional value, it is not always necessary to write it. But this rule is useful if you want to force the writing of description. Especially it is useful when combined with [terraform-docs](https://github.com/segmentio/terraform-docs).

## How To Fix

Write a description other than an empty string.
