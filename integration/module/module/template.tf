variable "unknown" {}
variable "enable" {}

variable "instance_type" {
  default = "t2.micro"
}

resource "aws_instance" "unknown" {
  instance_type = var.unknown
}

module "instance" {
  source = "./module"

  enable = var.enable
  instance_type = var.instance_type
}
