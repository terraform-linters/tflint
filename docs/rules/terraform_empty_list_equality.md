# terraform_empty_list_equality

Disallow comparisons with `[]` when checking if a collection is empty.

## Example

```hcl
variable "my_list" {
	type = list(string)
}
resource "aws_db_instance" "mysql" {
	count = var.my_list == [] ? 0 : 1
    instance_class = "m4.2xlarge"
}
```

```
$ tflint
1 issue(s) found:

Warning: Comparing a collection with an empty list is invalid. To detect an empty collection, check its length. (terraform_empty_list_equality)

  on test.tf line 5:
   5:   count = var.my_list == [] ? 0 : 1

Reference: https://github.com/terraform-linters/tflint/blob/master/docs/rules/terraform_empty_list_equality.md
 
```

## Why

The `==` operator can only return true when the two operands have identical types, and the type of `[]` alone (without any further type conversions) is an empty tuple rather than a list of objects, strings, numbers or any other type. Therefore, a comparison with a single `[]` with the goal of checking if a collection is empty, will always return false.

## How To Fix

Check if a collection is empty by checking its length instead. For example: `length(var.my_list) == 0`.
