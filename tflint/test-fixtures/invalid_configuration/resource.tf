resources "aws_instance" "web" {
  ami           = "ami-12345678"
  instance_type = "t1.2xlarge"
}
