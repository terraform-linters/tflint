# terraform_module_semver_source

Disallow specifying a Git or Mercurial repository as a module source without pinning to a semantic version reference.

## Example

```hcl
module "unpinned" {
  source = "git://hashicorp.com/consul.git"
}

module "git_commit_ref" {
  source = "git://hashicorp.com/consul.git?ref=8afd21a"
}

module "git_semver_ref" {
  source = "git://hashicorp.com/consul.git?ref=v1.0.0"
}

module "mercurial_branch_rev" {
  source = "hg::http://hashicorp.com/consul.hg?rev=branch"
}

module "mercurial_semver_rev" {
  source = "hg::http://hashicorp.com/consul.hg?rev=1.2.3"
}
```

```
$ tflint example.tf --disable-rule=terraform_module_pinned_source --enable-rule=terraform_module_semver_source
3 issue(s) found:

Warning: Module source "git://hashicorp.com/consul.git" is not pinned (terraform_module_semver_source)

  on example.tf line 2:
   2:   source = "git://hashicorp.com/consul.git"

Reference: https://github.com/wata727/tflint/blob/v0.12.1/docs/rules/terraform_module_semver_source.md

Warning: Module source "git://hashicorp.com/consul.git?ref=8afd21a" uses a ref which is not a version string (terraform_module_semver_source)

  on example.tf line 6:
   6:   source = "git://hashicorp.com/consul.git?ref=8afd21a"

Reference: https://github.com/wata727/tflint/blob/v0.12.1/docs/rules/terraform_module_semver_source.md

Warning: Module source "hg::http://hashicorp.com/consul.hg?rev=branch" uses a rev which is not a version string (terraform_module_semver_source)

  on example.tf line 14:
  14:   source = "hg::http://hashicorp.com/consul.hg?rev=branch"

Reference: https://github.com/wata727/tflint/blob/v0.12.1/docs/rules/terraform_module_semver_source.md
```

## Why

Terraform allows you to checkout module definitions from source control. If you specify a non-permanent reference (for example, a branch name), the dependency may change without your awareness. To prevent this, always specify an explicit version to checkout.

## How To Fix

Specify a semantic version pin, of the form `vX.Y.Z`.  The leading `v` is optional.
