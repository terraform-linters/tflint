resource "aws_instance" "foo" {
    instance_type = "${path.cwd}/instance_type"
}
