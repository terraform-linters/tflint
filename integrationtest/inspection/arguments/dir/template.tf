resource "aws_instance" "template" {
  instance_type = "t1.2xlarge"
}

// `terraform get` in Terraform v0.12
module "instances" {
  source = "./module"
}
