variable "zero" {
  default = 0
}

variable "one" {
  default = 1
}

variable "empty_object" {
  default = {}
}

variable "object" {
  default = {
    foo = "bar"
  }
}

variable "empty_set" {
  default = []
}

variable "set" {
  default = ["foo", "bar"]
}

variable "unknown" {}

resource "aws_instance" "zero" {
  count = var.zero
  instance_type = "t2.micro"
}

resource "aws_instance" "one" {
  count = var.one
  instance_type = "t2.micro"
}

resource "aws_instance" "unknown_count" {
  count = var.unknown
  instance_type = "t2.micro"
}

resource "aws_instance" "empty_object" {
  for_each = var.empty_object
  instance_type = "t2.micro"
}

resource "aws_instance" "object" {
  for_each = var.object
  instance_type = "t2.micro"
}

resource "aws_instance" "empty_set" {
  for_each = var.empty_set
  instance_type = "t2.micro"
}

resource "aws_instance" "set" {
  for_each = var.set
  instance_type = "t2.micro"
}

resource "aws_instance" "unknown_for_each" {
  for_each = var.unknown
  instance_type = "t2.micro"
}
