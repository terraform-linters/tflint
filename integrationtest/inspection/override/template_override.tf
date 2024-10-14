resource "aws_instance" "web" {
  ami                  = "ami-12345678"
  instance_type        = "m5.2xlarge" // aws_instance_example_type
  iam_instance_profile = "web-server"
}

terraform {
  required_providers {
    aws     = "2"
    google  = "2"
    oracle  = "2"
  }
}
