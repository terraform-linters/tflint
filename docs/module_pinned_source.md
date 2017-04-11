# Terraform Module Pinned Source
This issue is reported if you specify a git or mercurial repository without pinning to a non-default version.

## Example
```
module "unpinned" {
	source = "git://hashicorp.com/consul.git"
}

module "default git" {
	source = "git://hashicorp.com/consul.git?ref=master"
}

module "default mercurial" {
	source = "hg::http://hashicorp.com/consul.hg?rev=default"
}
```

The following is the execution result of TFLint: 

```
$ tflint
template.tf
				WARNING:2 Module source "git://hashicorp.com/consul.git" is not pinned
				WARNING:6 Module source "git://hashicorp.com/consul.git?ref=master" uses default ref "master"
				WARNING:10 Module source "hg::http://hashicorp.com/consul.hg?rev=default" uses default rev "default"

Result: 3 issues  (0 errors , 3 warnings , 0 notices)
```

## Why
Terraform allows you to checkout module definitions from source control. If you do not pin the version to checkout, the dependency you require may introduce major breaking changes without your awareness. To preven this, always specify an explicit version to checkout.

## How To Fix
Specify a version pin.  For git repositories, it should not be "master". For Mercurial repositories, it should not be "default"
