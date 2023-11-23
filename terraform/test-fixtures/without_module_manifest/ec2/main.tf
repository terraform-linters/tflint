variable "ami_id" {}

resource "aws_instance" "main" {
  ami           = var.ami_id
  instance_type = "t2.micro"
}
