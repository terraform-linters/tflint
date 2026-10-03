variable "instance_type" {}

locals {
  instance_type = var.instance_type
}

resource "aws_instance" "via_local" {
  // The attribute depends on `instance_type` through a local value
  instance_type = local.instance_type
}
