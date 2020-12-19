# terraform_unused_required_providers

Check that all `required_providers` are used in the module.

## Configuration

```hcl
rule "terraform_unused_required_providers" {
  enabled = true
}
```

## Examples

```hcl
terraform {
  required_providers {
    null = {
      source = "hashicorp/null"
    }
  }
}
```

```
$ tflint
1 issue(s) found:

Warning: provider 'null' is declared in required_providers but not used by the module (terraform_unused_required_providers)

  on main.tf line 3:
   3:     null = {
   4:       source = "hashicorp/null"
   5:     }

Reference: https://github.com/terraform-linters/tflint/blob/v0.22.0/docs/rules/terraform_unused_required_providers.md
```

## Why

The `required_providers` block should specify providers used directly by the given Terraform module. Terraform will download all specified providers during `terraform init`. If all resources for a given provider are removed but the `required_providers` entry remains, Terraform will continue to download the provider.

In general, each module should specify its own provider requirements for each provider it uses. Terraform will traverse the module graph and find a suitable version for all providers, or error if modules require conflicting versions. 

## How To Fix

If the provider is no longer used, remove it from the `required_providers` block. 

If the provider is used in one or more child modules but not directly in the module where TFLint was invoked, cut and paste the provider requirement into those modules.

If the provider is used in one or more child modules and you'd prefer to define a single requirement, you can ignore the warning:

```tf
terraform {
  required_providers {
    # tflint-ignore: terraform_unused_required_providers
    null = {
      source = "hashicorp/null"
    }
  }
}
```

This will affect your ability to run `terraform` directly in the child module, especially if you use providers outside the default `hashicorp` namespace or specify a `version` for required providers ([recommended](./terraform_required_providers.md)).
