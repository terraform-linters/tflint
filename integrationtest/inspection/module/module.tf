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

module "instances_for_each" {
  source = "./module"

  for_each = toset(["t1.4xlarge"])

  unknown = var.unknown
  enable = true
  instance_type = each.key
}

module "instances_with_annotations" {
  source = "./module"

  unknown = var.unknown
  // tflint-ignore: aws_instance_example_type
  enable = true
  // tflint-ignore: aws_instance_example_type
  instance_type = var.instance_type
}

module "ignored_instances" {
  source = "./ignore_module"

  instance_type = "t1.2xlarge"
}
