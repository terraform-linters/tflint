variable "tags" {
  default = []
}

resource "aws_autoscaling_group" "group" {
  tags = var.tags
}
