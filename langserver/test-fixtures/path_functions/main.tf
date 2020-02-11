resource "aws_instance" "foo" {
    instance_type = file("${path.cwd}/instance_type")
}
