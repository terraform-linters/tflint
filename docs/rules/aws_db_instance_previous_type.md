# aws_db_instance_previous_type

Disallow using previous generation instance types.

## Example

```hcl
resource "aws_db_instance" "default" {
  allocated_storage    = 10
  engine               = "mysql"
  engine_version       = "5.6.17"
  instance_class       = "db.t1.micro" // previous generation instance type!
  name                 = "mydb"
  username             = "foo"
  password             = "bar"
  db_subnet_group_name = "my_database_subnet_group"
  parameter_group_name = "default.mysql5.6"
}
```

```
$ tflint
1 issue(s) found:

Warning: "db.t1.micro" is previous generation instance type. (aws_db_instance_previous_type)

  on template.tf line 5:
   5:   instance_class       = "db.t1.micro" // previous generation instance type!

Reference: https://github.com/terraform-linters/tflint/blob/v0.11.0/docs/rules/aws_db_instance_previous_type.md
 
```

## Why

Current generation instance types have better performance and lower cost than previous generations. Users should avoid previous generation instance types, especially for new instances.

## How To Fix

Select a current generation instance type according to the [upgrade paths](https://aws.amazon.com/rds/previous-generation/).
