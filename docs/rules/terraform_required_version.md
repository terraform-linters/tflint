# terraform_required_version

Disallow `terraform` declarations without `required_version`.

## Configuration

```hcl
rule "terraform_required_version" {
  enabled = true
}
```

## Example

```hcl
terraform {
  required_version = ">= 1.0" 
}
```

```
$ tflint
1 issue(s) found:

Warning: terraform "required_version" attribute is required

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_required_version.md 
```

## Why
The `required_version` setting can be used to constrain which versions of the Terraform CLI can be used with your configuration. 
If the running version of Terraform doesn't match the constraints specified, Terraform will produce an error and exit without 
taking any further actions.

## How To Fix

Add the `required_version` attribute to the terraform configuration block.
