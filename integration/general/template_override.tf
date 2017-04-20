resource "aws_instance" "web" {
  ami                  = "ami-12345678"
  instance_type        = "t2.2xlarge"
  iam_instance_profile = "web-server"
}
