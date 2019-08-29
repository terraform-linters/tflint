variable "unknown" {}

variable "instance_type" {
  default = "t1.2xlarge"
}

module "instances" {
  source = "./module"

  unknown = var.unknown
  enable = true
  instance_type = var.instance_type
}

module "instances_with_annotations" {
  source = "./module"

  unknown = var.unknown
  // tflint-ignore: aws_instance_invalid_type
  enable = true
  // tflint-ignore: aws_instance_invalid_type
  instance_type = var.instance_type
}

module "ignored_instances" {
  source = "./ignore_module"

  instance_type = "t1.2xlarge"
}
