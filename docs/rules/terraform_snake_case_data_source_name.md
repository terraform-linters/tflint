# terraform_snake_case_data_source_name

Enforces snake_case for `data` source names. A warning will be thrown if there are any characters in the name that are not lower case or an underscore (`_`).

## Example

```hcl
data "aws_eip" "camelCase" {
}

data "aws_eip" "valid_name" {
}
```

```
$ tflint
1 issue(s) found:

Notice: `camelCase` data source name is not snake_case (terraform_snake_case_data_source_name)

  on template.tf line 1:
   1: data "aws_eip" "camelCase" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.15.3/docs/rules/terraform_snake_case_data_source_name.md
 
```

## Why

Naming conventions are optional, so it is not necessary to follow this. But this rule is useful if you want to force the following naming conventions in line with the [Terraform Plugin Naming Best Practices](https://www.terraform.io/docs/extend/best-practices/naming.html).

## How To Fix

Use lower case characters and separate words with underscores (`_`)
