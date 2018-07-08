resource "aws_instance" "invalid" {
  instance_type = "t1.2xlarge"
}

resource "aws_instance" "valid" {
  instance_type = "t2.micro"
}

resource "aws_instance" "missing_key" {
  ami = "ami-12345678"
}
