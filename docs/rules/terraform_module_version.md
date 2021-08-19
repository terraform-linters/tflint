# terraform_module_version

Ensure that all modules sourced from a [Terraform Registry](https://www.terraform.io/docs/language/modules/sources.html#terraform-registry) specify a `version`.

## Configuration

Name | Description | Default | Type
--- | --- | --- | ---
exact | Require an exact version | false | Boolean

```hcl
rule "terraform_module_version" {
  enabled = true
  exact = false # default
}
```

## Example

```tf
module "exact" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "1.0.0"
}

module "range" {
  source  = "terraform-aws-modules/vpc/aws"
  version = ">= 1.0.0"
}

module "latest" {
  source  = "terraform-aws-modules/vpc/aws"
}
```

```
$ tflint
1 issue(s) found:

Warning: module "latest" should specify a version (terraform_module_version)

  on main.tf line 11:
  11: module "latest" {

Reference: https://github.com/terraform-linters/tflint/blob/master/docs/rules/terraform_module_version.md
```

### Exact

```hcl
rule "terraform_module_version" {
  enabled = true
  exact = true
}
```

```tf
module "exact" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "1.0.0"
}

module "range" {
  source  = "terraform-aws-modules/vpc/aws"
  version = ">= 1.0.0"
}
```

```
$ tflint
1 issue(s) found:

Warning: module "range" should specify an exact version, but a range was found (terraform_module_version)

  on main.tf line 8:
   8:   version = ">= 1.0.0"

Reference: https://github.com/terraform-linters/tflint/blob/master/docs/rules/terraform_module_version.md
```

## Why

Terraform's [module version documentation](https://www.terraform.io/docs/language/modules/syntax.html#version) states:

> When using modules installed from a module registry, we recommend explicitly constraining the acceptable version numbers to avoid unexpected or unwanted changes.

When no `version` is specified, Terraform will download the latest version available on the registry. Using a new major version of a module could cause the destruction of existing resources, or the creation of new resources that are not backwards compatible. Generally you should at least constrain modules to a specific major version.

### Exact Versions

Depending on your workflow, you may want to enforce that modules specify an _exact_ version by settings `exact = true` for this rule. This will disallow any module that includes multiple comma-separated version constraints, or any [constraint operator](https://www.terraform.io/docs/language/expressions/version-constraints.html#version-constraint-syntax) other than `=`. Exact versions are often used with automated dependency managers like [Dependabot](https://docs.github.com/en/code-security/supply-chain-security/keeping-your-dependencies-updated-automatically/about-dependabot-version-updates) and [Renovate](https://docs.renovatebot.com), which will automatically propose a pull request to update the module when a new version is released.

Keep in mind that the module may include further child modules, which have their own version constraints. TFLint _does not_ check version constraints set in child modules. **Enabling this rule cannot guarantee that `terraform init` will be deterministic**. Use [Terraform dependency lock files](https://www.terraform.io/docs/language/dependency-lock.html) to ensure that Terraform will always use the same version of all modules (and providers) until you explicitly update them.

## How To Fix

Specify a `version`. If `exact = true`, this must be an exact version.
