resource "aws_instance" "foo" {
    // tflint-ignore: aws_instance_invalid_instance_type
    instance_type = "t2.micro"
}
