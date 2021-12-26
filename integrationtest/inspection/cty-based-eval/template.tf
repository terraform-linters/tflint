variable "tags" {
  default = [
    { key: "foo", value: "bar", propagate_at_launch: true }
  ]
}

resource "aws_autoscaling_group" "foo" {
  tags = var.tags
}
