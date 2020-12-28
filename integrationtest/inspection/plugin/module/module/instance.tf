variable "instance_type" {}

variable "enable" {
  default = false
}

resource "aws_instance" "dependent" {
  // The attribute depends on `enable` and `instance_type`
  instance_type = var.enable ? var.instance_type : "t2.micro"
}

resource "aws_instance" "independent" {
  // instance_type is invalid, but the attribute does not depend on the parent module aruguments
  instance_type = "t1.2xlarge"
}
