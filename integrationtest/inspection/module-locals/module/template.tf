variable "instance_type" {}

locals {
  instance_type = var.instance_type
}

module "instance" {
  source = "./module"

  // The argument depends on `instance_type` through a local value
  instance_type = local.instance_type
}
