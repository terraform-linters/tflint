# terraform_module_pinned_source

Disallow specifying a git or mercurial repository as a module source without pinning to a version.

## Configuration

Name | Default | Value
--- | --- | ---
enabled | true | Boolean
style | `flexible` | `flexible`, `semver`

```hcl
rule "terraform_module_pinned_source" {
  enabled = true
  style = "flexible"
}
```

## Example

### style = "flexible"

In the "flexible" style, all sources must be pinned to non-default version.

```hcl
module "unpinned" {
  source = "git://hashicorp.com/consul.git"
}

module "default_git" {
  source = "git://hashicorp.com/consul.git?ref=master"
}

module "default_mercurial" {
  source = "hg::http://hashicorp.com/consul.hg?rev=default"
}

module "pinned_git" {
  source = "git://hashicorp.com/consul.git?ref=feature"
}
```

```
$ tflint
3 issue(s) found:

Warning: Module source "git://hashicorp.com/consul.git" is not pinned (terraform_module_pinned_source)

  on template.tf line 2:
   2:   source = "git://hashicorp.com/consul.git"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md

Warning: Module source "git://hashicorp.com/consul.git?ref=master" uses default ref "master" (terraform_module_pinned_source)

  on template.tf line 6:
   6:   source = "git://hashicorp.com/consul.git?ref=master"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md

Warning: Module source "hg::http://hashicorp.com/consul.hg?rev=default" uses default rev "default" (terraform_module_pinned_source)

  on template.tf line 10:
  10:   source = "hg::http://hashicorp.com/consul.hg?rev=default"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md

```

### style = "semver"

In the "semver" style, all sources must be pinned to semantic version reference. This is stricter than the "flexible" style.

```hcl
module "unpinned" {
  source = "git://hashicorp.com/consul.git"
}

module "pinned_to_branch" {
  source = "git://hashicorp.com/consul.git?ref=feature"
}

module "pinned_to_version" {
  source = "git://hashicorp.com/consul.git?ref=v1.2.0"
}
```

```
$ tflint
2 issue(s) found:

Warning: Module source "git://hashicorp.com/consul.git" is not pinned (terraform_module_pinned_source)

  on template.tf line 2:
   2:   source = "git://hashicorp.com/consul.git"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md

Warning: Module source "git://hashicorp.com/consul.git?ref=feature" uses a ref which is not a version string (terraform_module_pinned_source)

  on template.tf line 6:
   6:   source = "git://hashicorp.com/consul.git?ref=feature"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md

```

## Why

Terraform allows you to checkout module definitions from source control. If you do not pin the version to checkout, the dependency you require may introduce major breaking changes without your awareness. To prevent this, always specify an explicit version to checkout.

More strictly, pinning to a non-permanent reference, such as a branch name, includes ambiguity. The "semver" style is used to avoid such cases.

## How To Fix

Specify a version pin.  For git repositories, it should not be "master". For Mercurial repositories, it should not be "default".

In the "semver" style: Specify a semantic version pin, of the form `vX.Y.Z`.  The leading `v` is optional.
