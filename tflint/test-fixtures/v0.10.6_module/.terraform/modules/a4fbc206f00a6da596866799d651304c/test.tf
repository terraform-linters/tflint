resource "aws_instance" "web" {
  ami           = "ami-abcd1234"
  instance_type = "t2.micro"
}
