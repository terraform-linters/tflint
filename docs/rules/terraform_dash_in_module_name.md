# terraform_dash_in_module_name

Disallow dashes (-) in `module` names.

## Example

```hcl
module dash-name" {
}

module "no_dash_name" {
}
```

```
$ tflint
1 issue(s) found:

Notice: `dash-name` module name has a dash (terraform_dash_in_module_name)

  on template.tf line 1:
   1: module "dash-name" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_dash_in_module_name.md

```

## Why

Naming conventions are optional, so it is not necessary to follow this. But this rule is useful if you want to force the following naming conventions in line with the [Terraform Plugin Naming Best Practices](https://www.terraform.io/docs/extend/best-practices/naming.html).

## How To Fix

Use underscores (_) instead of dashes (-).
