# terraform_deprecated_index

Disallow legacy dot index syntax.

## Example

```hcl
locals {
  list  = ["a", "b", "c"]
  value = list.0 
}
```

```
$ tflint
1 issue(s) found:

Warning: List items should be accessed using square brackets (terraform_deprecated_index)

  on example.tf line 3:
   3:   value = list.0

Reference: https://github.com/terraform-linters/tflint/blob/v0.16.1/docs/rules/terraform_deprecated_index.md
```

## Why

Terraform v0.12 supports traditional square brackets for accessing list items by index. However, for backward compatability with v0.11, Terraform continues to support accessing list items with the dot syntax normally used for attributes. While Terraform does not print warnings for this syntax, it is no longer documented and its use is discouraged.

## How To Fix

Switch to the square bracket syntax when accessing items in list, including resources that use `count`.
