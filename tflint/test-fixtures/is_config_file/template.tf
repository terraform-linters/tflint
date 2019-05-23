resource "aws_instance" "web" {
  ami           = "ami-b73b63a0"
  instance_type = "t1.2xlarge" # invalid type!

  tags {
    Name = "HelloWorld"
  }
}
