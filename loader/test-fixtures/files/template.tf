provider "aws" {
    region = "us-east-1"
}

resource "aws_instance" "web" {
    ami = "ami-12345"
    instance_type = "t2.micro"
}