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
template.tf
        WARNING:5 "db.t1.micro" is previous generation instance type. (aws_db_instance_previous_type)

Result: 1 issues  (0 errors , 1 warnings , 0 notices)
```

## Why

Previous generation instance types are inferior to current generation in terms of performance and fee. Unless there is a special reason, you should avoid to use these ones.

## How To Fix

Select a current generation instance type according to the [upgrade paths](https://aws.amazon.com/rds/previous-generation/).
