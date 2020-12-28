module "ec2" {
  source = "./module"

  root_suffix = "t2.micro"
  module_suffix = "t3.micro"
}
