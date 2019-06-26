resource "aws_instance" "web" {
  ami                  = "ami-12345678"
  instance_type        = "t1.1xlarge" // aws_instance_invalid_instance_type
  iam_instance_profile = "web-server"
}
