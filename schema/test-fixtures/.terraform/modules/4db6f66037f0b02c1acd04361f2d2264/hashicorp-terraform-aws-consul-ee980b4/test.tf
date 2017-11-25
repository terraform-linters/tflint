resource "aws_instance" "web" {
  ami           = "ami-76fba12"
  instance_type = "c5.2xlarge"
}
