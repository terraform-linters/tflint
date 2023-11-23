module "instance" {
  source = "./ec2"
  ami_id = "ami-1234abcd"
}

module "consul" {
   source = "hashicorp/consul/aws"
   version = "0.9.0"
}
