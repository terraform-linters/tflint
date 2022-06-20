variable "instance_type" {}

resource "aws_instance" "foo" {
  ami = "ami-12345678"
  instance_type = var.instance_type
}
