# terraform_required_providers

Require that all providers have version constraints through `required_providers` or the provider `version` attribute.

## Configuration

```hcl
rule "terraform_required_providers" {
  enabled = true
}
```

## Example

```hcl
provider "template" {}
```

```
$ tflint
1 issue(s) found:

Warning: Provider "template" should have a version constraint in required_providers

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_required_providers.md 
```

## Why

Providers are plugins released on a separate rhythm from Terraform itself, and so they have their own version numbers. For production use, you should constrain the acceptable provider versions via configuration, to ensure that new versions with breaking changes will not be automatically installed by `terraform init` in future.

## How To Fix

Add the `required_providers` attribute to the `terraform` configuration block and include current versions for all providers. For example:

```tf
terraform {
  required_providers {
    template = "~> 2.0"
  }
}
```

Provider version constraints can also be specified using a version argument within a provider block but this is not recommend, particularly for child modules.
