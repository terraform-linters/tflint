variable "mysql_types" {
  type = "list"
}

resource "aws_db_instance" "mysql" {
  allocated_storage    = 10
  engine               = "mysql"
  engine_version       = "5.6.17"
  instance_class       = "${var.mysql_types[0]}}"   // aws_db_instance_invalid_type
  name                 = "mydb"
  username             = "foo"
  password             = "secret_password"          // aws_db_instance_readable_password
  db_subnet_group_name = "my_database_subnet_group"
  parameter_group_name = "default.mysql5.6"         // aws_db_instance_default_parameter_group
}

resource "aws_db_instance" "app" {
  allocated_storage    = 10
  engine               = "mysql"
  engine_version       = "5.6.17"
  instance_class       = "${var.mysql_types[1]}"    // aws_db_instance_previous_type
  name                 = "mydb"
  username             = "foo"
  password             = "secret_password"          // aws_db_instance_readable_password
  db_subnet_group_name = "my_database_subnet_group"
  parameter_group_name = "application.mysql5.6"
}
