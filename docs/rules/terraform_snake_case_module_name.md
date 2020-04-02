# terraform_snake_case_module_name

Enforces snake_case for `module` names. A warning will be thrown if there are any characters in the name that are not lower case or an underscore (`_`).

## Example

```hcl
module "camelCase" {
}

module "valid_name" {
}
```

```
$ tflint
1 issue(s) found:

Notice: `camelCase` module name is not snake_case (terraform_snake_case_module_name)

  on template.tf line 1:
   1: module "camelCase" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.3/docs/rules/terraform_snake_case_module_name.md
 
```

## Why

Naming conventions are optional, so it is not necessary to follow this. But this rule is useful if you want to force the following naming conventions in line with the [Terraform Plugin Naming Best Practices](https://www.terraform.io/docs/extend/best-practices/naming.html).

## How To Fix

Use lower case characters and separate words with underscores (`_`)
