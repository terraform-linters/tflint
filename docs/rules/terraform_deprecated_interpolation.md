# terraform_deprecated_interpolation

Disallow deprecated (0.11-style) interpolation

## Example

```hcl
resource "aws_instance" "deprecated" {
    instance_type = "${var.type}"
}

resource "aws_instance" "new" {
    instance_type = var.type
}
```

```
$ tflint
1 issue(s) found:

Warning: Interpolation-only expressions are deprecated in Terraform v0.12.14 (terraform_deprecated_interpolation)

  on example.tf line 2:
   2:     instance_type = "${var.type}"

Reference: https://github.com/terraform-linters/tflint/blob/v0.14.0/docs/rules/terraform_deprecated_interpolation.md
 
```

## Why

Terraform v0.12 introduces a new interpolation syntax, but continues to support the old 0.11-style interpolation syntax for compatibility.

Terraform will currently print diagnostic warnings when deprecated interpolations are used. Consistent with its deprecation policy, they will raise errors in the next major release (v0.13). TFLint emits an issue instead of a warning with the same logic.

## How To Fix

Switch to the new interpolation syntax. See the release notes for Terraform 0.12.14 for details: https://github.com/hashicorp/terraform/releases/tag/v0.12.14