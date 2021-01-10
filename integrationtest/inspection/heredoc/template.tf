resource "aws_instance" "foo" {
  instance_type = <<EOF
t2.micro
EOF
}
