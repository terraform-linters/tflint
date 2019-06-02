resource "aws_instance" "backend" {
  ami           = "ami-b73b63a0"
  instance_type = "t1.2xlarge"

  tags {
    Name = "HelloWorld"
  }
}
