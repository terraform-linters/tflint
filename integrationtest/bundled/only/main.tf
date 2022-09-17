variable "unused" {}
variable "used" {}

resource "aws_instance" "main" {
  instance_type = "${var.used}"
}
