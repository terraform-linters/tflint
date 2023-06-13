resource "aws_instance" "autofixed_literal" {
  instance_type = "[AUTO_FIXED]"
}

module "instances" {
  source = "./module"

  input = "[AUTO_FIXED]"
}
