variable "unknown" {}

variable "instance_type" {
  default = "t1.2xlarge"
}

resource "aws_instance" "foo" {
  instance_type = var.instance_type
}

module "instances" {
  source = "./module"

  unknown = var.unknown
  // tflint-ignore: aws_instance_example_type
  enable = true
  instance_type = var.instance_type
}
