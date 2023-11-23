variable "instance_type" {}

resource "aws_instance" "main" {
  instance_type = var.instance_type
}
