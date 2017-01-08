# AWS DB Instance Previous Type
Report this issue if you have specified the previous instance type. This issue type is WARNING.

## Example
```
resource "aws_db_instance" "default" {
  allocated_storage    = 10
  engine               = "mysql"
  engine_version       = "5.6.17"
  instance_class       = "db.t1.micro" // previous type!
  name                 = "mydb"
  username             = "foo"
  password             = "bar"
  db_subnet_group_name = "my_database_subnet_group"
  parameter_group_name = "default.mysql5.6"
}
```

The following is the execution result of TFLint:

```
$ tflint
template.tf
        WARNING:5 "db.t1.micro" is previous generation instance type.

Result: 1 issues  (0 errors , 1 warnings , 0 notices)
```

## Why
There are two types of instance types, the current generation and the previous generation. The current generation is superior to the previous generation in terms of performance and fee. AWS also officially states that unless there is a special reason, you should use the instance type of the current generation.

## How to fix
Follow the [upgrade paths](https://aws.amazon.com/jp/rds/previous-generation/) and confirm that the instance type of the current generation can be used, then select again.
