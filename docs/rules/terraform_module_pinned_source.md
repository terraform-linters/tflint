# terraform_module_pinned_source

Disallow specifying a git or mercurial repository as a module source without pinning to a version.

## Configuration

Name | Default | Value
--- | --- | ---
enabled | true | Boolean
style | `flexible` | `flexible`, `semver`
default_branches | `["master", "main", "default", "develop"]` | 

```hcl
rule "terraform_module_pinned_source" {
  enabled = true
  style = "flexible"
  default_branches = ["dev"]
}
```

Configured `default_branches` will be appended to the defaults rather than overriding them.

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

Warning: Module source "git://hashicorp.com/consul.git?ref=master" uses a default branch as ref (master) (terraform_module_pinned_source)

  on template.tf line 6:
   6:   source = "git://hashicorp.com/consul.git?ref=master"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md

Warning: Module source "hg::http://hashicorp.com/consul.hg?rev=default" uses a default branch as rev (default) (terraform_module_pinned_source)

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

Warning: Module source "git://hashicorp.com/consul.git?ref=feature" uses a ref which is not a semantic version string (terraform_module_pinned_source)

  on template.tf line 6:
   6:   source = "git://hashicorp.com/consul.git?ref=feature"

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.0/docs/rules/terraform_module_pinned_source.md

```

## Why

Terraform allows you to source modules from source control repositories. If you do not pin the revision to use, the dependency you require may introduce unexpected breaking changes. To prevent this, always specify an explicit version to check out.

Pinning to a mutable reference, such as a branch, still allows for unintended breaking changes. Semver style can help avoid this.

## How To Fix

Specify a version pin.  For git repositories, it should not be "master". For Mercurial repositories, it should not be "default".

In the "semver" style: specify a semantic version pin of the form `vX.Y.Z`. The leading `v` is optional.
