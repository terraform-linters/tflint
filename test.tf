variable "my_list" {
	type = list(string)
}
resource "aws_db_instance" "mysql" {
	count = var.my_list == [] ? 0 : 1
    instance_class = "m4.2xlarge"
}
