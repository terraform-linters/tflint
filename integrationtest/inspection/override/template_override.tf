resource "aws_instance" "web" {
  ami                  = "ami-12345678"
  instance_type        = "m5.2xlarge" // aws_instance_example_type
  iam_instance_profile = "web-server"
}
