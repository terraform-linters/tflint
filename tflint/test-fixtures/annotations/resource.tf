resource "aws_instance" "foo" {
    /* tflint-ignore: aws_instance_invalid_type */
    instance_type = "t2.micro" // tflint-ignore: aws_instance_invalid_type
    # tflint-ignore: aws_instance_invalid_type This is also comment
    iam_instance_profile = "foo" # This is also comment
    // This is also comment
}
