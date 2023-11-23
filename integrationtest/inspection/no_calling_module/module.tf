variable "instance_type" {
  default = "t1.2xlarge"
}

resource "aws_instance" "main" {
  instance_type = var.instance_type
}

module "local" {
  source = "./ec2-instance"

  instance_type = var.instance_type
}

module "remote" {
  source = "terraform-aws-modules/ec2-instance/aws"

  instance_type = var.instance_type
}
