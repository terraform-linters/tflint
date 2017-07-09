# AWS DB Instance Default Parameter Group
Report this issue if you have specified the default parameter group. This issue type is NOTICE.

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
  vpc_security_group_ids = ["${aws_security_group.mysql.id}"]
  db_subnet_group_name   = "app-subnet-group"
  parameter_group_name   = "default.mysql5.7"
  multi_az               = true
}
```

The following is the execution result of TFLint: 

```
$ tflint
template.tf
        NOTICE:12 "default.mysql5.7" is default parameter group. You cannot edit it. (aws_db_instance_default_parameter_group)

Result: 1 issues  (0 errors , 0 warnings , 1 notices)
```

## Why
RDS allows you to use a parameter group structure to change setting values ​​such as MySQL and PostgreSQL. However, the default parameter group can not be changed later, and if you want to change it, you need to create and modify a dedicated parameter group.

## How To Fix
Please create a dedicated parameter group and change it to that.
