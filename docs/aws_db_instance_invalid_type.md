# AWS DB Instance Invalid Type
Report this issue if you have specified the invalid instance type. This issue type is ERROR.

## Example
```
resource "aws_db_instance" "default" {
  allocated_storage    = 10
  engine               = "mysql"
  engine_version       = "5.6.17"
  instance_class       = "t1.micro" // invalid type!
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
        ERROR:5 "t1.micro" is invalid instance type.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
If an invalid instance type is specified, an error will occur at `terraform apply`.

## How To Fix
Check the [instance type list](https://aws.amazon.com/rds/details/) and select a valid instance type again.
