variable "input" {}

resource "aws_instance" "main" {
  instance_type = var.input
}
