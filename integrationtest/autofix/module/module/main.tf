variable "input" {}

resource "aws_instance" "autofixed_literal" {
  instance_type = "[AUTO_FIXED]"
}

resource "aws_instance" "autofixed_variable" {
  instance_type = var.input
}
