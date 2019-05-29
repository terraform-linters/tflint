resource "aws_instance" "bar" {
    instance_type = "t2.micro" // tflint-ignore: aws_instance_invalid_type
}
