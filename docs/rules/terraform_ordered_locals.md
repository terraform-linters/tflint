# terraform_ordered_locals

Recommend proper order for variables in `locals` blocks.

Those variables are sorted based on their names (alphabetic order).

## Example

```hcl
locals {
  service_name = "forum"
  owner        = "Community Team"
}
```

```
$ tflint
1 issue(s) found:

Notice: Local values must be in alphabetical order (terraform_ordered_locals)

  on main.tf line 1:
   1: locals {

Reference: https://github.com/terraform-linters/tflint-ruleset-terraform/blob/v0.1.0/docs/rules/terraform_ordered_locals.md
```

## Why

It helps to improve the readability of terraform code by sorting variables in `locals` blocks in the order above.

## How To Fix

Sort variables in `locals` block in alphabetic order.