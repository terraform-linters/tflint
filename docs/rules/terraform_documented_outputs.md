# terraform_documented_outputs

Disallow `output` declarations without description.

## Example

```hcl
output "no_description" {
  value = "value"
}

output "empty_description" {
  value = "value"
  description = ""
}

output "description" {
  value = "value"
  description = "This is description"
}
```

```
$ tflint
outputs.tf
        NOTICE:1 `no_description` output has no description (terraform_documented_outputs)
        NOTICE:5 `empty_description` output has no description (terraform_documented_outputs)
```

## Why

Since `description` is optional value, it is not always necessary to write it. But this rule is useful if you want to force the writing of description. Especially it is useful when combined with [terraform-docs](https://github.com/segmentio/terraform-docs).

## How To Fix

Write a description other than an empty string.
