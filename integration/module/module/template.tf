variable "unknown" {}

variable "instance_type" {
  default = "t2.micro"
}

resource "aws_instance" "unknown" {
  instance_type = "${var.unknown}"
}

resource "aws_instance" "instance_type" {
  instance_type = "${var.instance_type}"
}
