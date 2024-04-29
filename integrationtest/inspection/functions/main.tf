resource "aws_instance" "core" {
  instance_type = upper("hello")
}

resource "aws_instance" "core_with_namespace" {
  instance_type = core::upper("hello")
}

resource "aws_instance" "provider" {
  instance_type = provider::tflint::instance_type()
}
