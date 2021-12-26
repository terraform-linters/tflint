resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}

resource "aws_s3_bucket" "foo" {
  bucket = "foo"
}
