resource "aws_instance" "intance" {
  tags = { foo = "bar" }
}

variable "sensitive" {
  sensitive = true
  default   = "sensitive"
}

resource "aws_instance" "sensitive" {
  tags = { foo = var.sensitive }
}
