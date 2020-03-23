# terraform_dash_in_data_source_name

Disallow dashes (-) in `data` source names.

## Example

```hcl
data "aws_eip" "dash-name" {
}

data "aws_eip" "no_dash_name" {
}
```

```
$ tflint
1 issue(s) found:

Notice: `dash-name` data source name has a dash (terraform_dash_in_data_source_name)

  on template.tf line 1:
   1: data "aws_eip" "dash-name" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/terraform_dash_in_data_source_name.md

```

## Why

Naming conventions are optional, so it is not necessary to follow this. But this rule is useful if you want to force the following naming conventions in line with the [Terraform Plugin Naming Best Practices](https://www.terraform.io/docs/extend/best-practices/naming.html).

## How To Fix

Use underscores (_) instead of dashes (-).
