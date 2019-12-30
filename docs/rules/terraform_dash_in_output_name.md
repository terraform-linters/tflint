# terraform_dash_in_output_name

Disallow dashes (-) in `output` names.

## Example

```hcl
output "dash-name" {
  value = "foo"
}

output "no_dash_name" {
  value = "foo"
}

```

```
$ tflint
1 issue(s) found:

Notice: `dash-bois-he` output name has a dash (terraform_dash_in_output_name)

  on outputs.tf line 1:
   1: output "dash-bois-he" {

Reference: https://github.com/terraform-linters/tflint/blob/master/docs/rules/terraform_dash_in_output_name.md
```

## Why

Naming conventions are optional, so it is not necessary to follow this. 

Whilst [Terraform Plugin Naming Best Practices](https://www.terraform.io/docs/extend/best-practices/naming.html) 
does not formally address outputs, I personally believe they are covered by this rule.


## How To Fix

Use underscores (_) instead of dashes (-).
