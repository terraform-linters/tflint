resource "aws_instance" "web" {
  ami           = "ami-8a7b6d5c"
  instance_type = "t2.nano"
}
