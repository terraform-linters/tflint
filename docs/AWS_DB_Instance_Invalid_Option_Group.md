# AWS DB Instance Invalid Option Group
Report this issue if you have specified the invalid option group name. This issue type is ERROR. This issue is enable only with deep check.

## Example
```
resource "aws_db_instance" "mysql" {
  identifier             = "app"
  allocated_storage      = 50
  storage_type           = "gp2"
  engine                 = "mysql"
  engine_version         = "5.7.11"
  instance_class         = "db.m4.large"
  name                   = "app_db"
  port                   = 3306
  publicly_accessible    = false
  vpc_security_group_ids = ["sg-12345678"]
  db_subnet_group_name   = "app-subnet-group"
  parameter_group_name   = "default.mysql5.7"
  option_group_name      = "invalid_option"
  multi_az               = true
}
```

The following is the execution result of TFLint: 

```
$ tflint --deep
template.tf
        ERROR:14 "invalid_option" is invalid option group name.

Result: 1 issues  (1 errors , 0 warnings , 0 notices)
```

## Why
If an invalid option group name is specified, an error will occur at `terraform apply`.

## How to fix
Check your option groups and select a valid name again.
