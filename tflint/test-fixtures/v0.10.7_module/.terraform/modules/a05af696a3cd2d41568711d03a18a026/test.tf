resource "aws_instance" "web" {
  ami           = "ami-9876abcd"
  instance_type = "m3.large"
}
