# terraform_required_version

Disallow `terraform` declarations without require_version.

## Configuration

Name | Default | Value
--- | --- | ---
enabled | true | Boolean
version |  | 

If a version is specified, the rule will ensure that the `required_version` matches the `version` of the rule.
```hcl
rule "terraform_required_version" {
  enabled = true
  version = "~> 0.12"
}
```

## Example

```hcl
terraform {
  required_providers {
    aws = ">= 2.7.0"
  }
}
```

```
$ tflint
1 issue(s) found:

Notice: terraform "required_version" attribute is required

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_required_version.md 
```

### version = "~> 0.12"

```hcl
terraform {
  required_version = "~> 0.11"
}
```

```
$ tflint
1 issue(s) found:

Notice: terraform "required_version" does not match specified version "~> 0.12"

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_required_version.md
 
```

## Why
The `required_version` setting can be used to constrain which versions of the Terraform CLI can be used with your configuration. 
If the running version of Terraform doesn't match the constraints specified, Terraform will produce an error and exit without 
taking any further actions.

## How To Fix

Add the `required_version` attribute to the terraform block.
