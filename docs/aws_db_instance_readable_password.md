# AWS DB Instance Readable Password
Report this issue if you write password for the master DB user directly. This issue type is WARNING.

## Example
```
resource "aws_db_instance" "default" {
  allocated_storage    = 10
  engine               = "mysql"
  engine_version       = "5.6.17"
  instance_class       = "db.t1.micro"
  name                 = "mydb"
  username             = "foo"
  password             = "jk4wu0o7" // readable password!
  db_subnet_group_name = "my_database_subnet_group"
  parameter_group_name = "default.mysql5.6"
}
```

The following is the execution result of TFLint:

```
$ tflint
template.tf
        WARNING:8 Password for the master DB user is readable. recommend using environment variables.

Result: 1 issues  (0 errors , 1 warnings , 0 notices)
```

Also, detect the following case:

```
variable "password" {
  description = "Password for MySQL master user"
  default     = "jk4wu0o7" // readable passowrd!
}

resource "aws_db_instance" "default" {
  allocated_storage    = 10
  engine               = "mysql"
  engine_version       = "5.6.17"
  instance_class       = "db.t1.micro"
  name                 = "mydb"
  username             = "foo"
  password             = "${var.password}"
  db_subnet_group_name = "my_database_subnet_group"
  parameter_group_name = "default.mysql5.6"
}
```


## Why
Generally, it is a bad practice to directly embed passwords in source code and templates. One reason why is that there is a fear that it will be unintentionally published when using VCS.

## How to fix
Instead of writing password directly, use environment variables. Terraform provides a way to set variables by environment variables. For example, edit and execute as following:

```
variable "password" {}

resource "aws_db_instance" "default" {
  allocated_storage    = 10
  engine               = "mysql"
  engine_version       = "5.6.17"
  instance_class       = "db.t1.micro"
  name                 = "mydb"
  username             = "foo"
  password             = "${var.password}"
  db_subnet_group_name = "my_database_subnet_group"
  parameter_group_name = "default.mysql5.6"
}
```

```
$ TF_VAR_password=jk4wu0o7 terraform apply
```

In the above case, The password cannot be read from templates. For details on how to set variables, please see the [documentation](https://www.terraform.io/intro/getting-started/variables.html).

NOTE: Unfortunately, even if you delete the password from templates, it will be stored in the state file. We recommend that encrypt state file and ignore that on VCS.

