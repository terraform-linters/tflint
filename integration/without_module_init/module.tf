variable "unknown" {}

variable "instance_type" {
  default = "t1.2xlarge"
}

// terraform init did not run
module "instances" {
  source = "./module"

  unknown = var.unknown
  instance_type = var.instance_type
}
