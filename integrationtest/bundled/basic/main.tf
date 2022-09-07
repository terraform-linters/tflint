variable "instance_type" {}
variable "unused" {
  type = string
}

resource "aws_instance" "main" {
  count = [] == [] ? 1 : 0

  instance_type = "${var.instance_type}"
}
