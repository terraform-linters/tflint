# terraform_naming_this

Enforces that resources are named "this" if only a single resource of this type is present.

This rule is based on point 2 from the [Terraform Best Practices for naming](https://www.terraform-best-practices.com/naming)

## Example

```hcl
resource "aws_s3_bucket" "wrong_name" {
  bucket = "test-bucket"
}
```

```
$ tflint
1 issue(s) found:

Notice: Found only one resource of type `aws_s3_bucket`, therefore the resource name should be `this` but was `wrong_name` (terraform_naming_this)

  on main.tf line 1:
   1: resource "aws_s3_bucket" "wrong_name" {

Reference: https://github.com/terraform-linters/tflint/blob/v0.23.1/docs/rules/terraform_naming_this.md

```

## How To Fix

Change the resource name to "this".

```hcl
resource "aws_s3_bucket" "this" {
  bucket = "test-bucket"
}
```
