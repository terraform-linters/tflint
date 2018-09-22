variable "unknown" {}

variable "instance_type" {
  default = "t1.2xlarge"
}

// `terraform init` in Terraform v0.11.8
module "instances" {
  source = "./module"

  unknown = "${var.unknown}"
  instance_type = "${var.instance_type}"
}
