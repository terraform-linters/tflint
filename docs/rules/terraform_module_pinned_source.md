# terraform_module_pinned_source

Disallow specifying a git or mercurial repository as a module source without pinning to a non-default version.

## Example

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
```

```
$ tflint
3 issue(s) found:

Warning: Module source "git://hashicorp.com/consul.git" is not pinned (terraform_module_pinned_source)

  on template.tf line 2:
   2:   source = "git://hashicorp.com/consul.git"

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_module_pinned_source.md

Warning: Module source "git://hashicorp.com/consul.git?ref=master" uses default ref "master" (terraform_module_pinned_source)

  on template.tf line 6:
   6:   source = "git://hashicorp.com/consul.git?ref=master"

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_module_pinned_source.md

Warning: Module source "hg::http://hashicorp.com/consul.hg?rev=default" uses default rev "default" (terraform_module_pinned_source)

  on template.tf line 10:
  10:   source = "hg::http://hashicorp.com/consul.hg?rev=default"

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_module_pinned_source.md
 
```

## Why

Terraform allows you to checkout module definitions from source control. If you do not pin the version to checkout, the dependency you require may introduce major breaking changes without your awareness. To prevent this, always specify an explicit version to checkout.

## How To Fix

Specify a version pin.  For git repositories, it should not be "master". For Mercurial repositories, it should not be "default"
