variable "unknown" {}

variable "default" {
  default = "default"
}

variable "default_values_file" {
  default = "default"
}

variable "auto_values_file" {
  default = "default"
}

variable "values_file" {
  default = "default"
}

variable "var" {
  default = "default"
}

resource "aws_instance" "unknown" {
  instance_type = var.unknown
}

resource "aws_instance" "default" {
  instance_type = var.default
}

resource "aws_instance" "default_values_file" {
  instance_type = var.default_values_file
}

resource "aws_instance" "auto_values_file" {
  instance_type = var.auto_values_file
}

resource "aws_instance" "values_file" {
  instance_type = var.values_file
}

resource "aws_instance" "var" {
  instance_type = var.var
}
