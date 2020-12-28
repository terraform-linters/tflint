variable "instance_type" {}

resource "aws_instance" "foo" {
  instance_type = var.instance_type
}
