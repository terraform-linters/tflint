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

```console
$ tflint
1 issue(s) found:

Warning: "db.t1.micro" is previous generation instance type. (aws_db_instance_previous_type)

  on template.tf line 5:
   5:   instance_class       = "db.t1.micro" // previous generation instance type!

Reference: https://github.com/wata727/tflint/blob/v0.11.0/docs/rules/aws_db_instance_previous_type.md

```

## Why

Previous generation instance types are inferior to current generation in terms of performance and fee. Unless there is a special reason, you should avoid to use these ones.

## How To Fix

Select a current generation instance type according to the [upgrade paths](https://aws.amazon.com/rds/previous-generation/).
