variable "auto_default" {
  default = "default"
}

variable "cli" {
  default = "default"
}

variable "auto" {
  default = "default"
}

variable "config" {
  default = "default"
}

resource "aws_instance" "foo" {
  instance_type = "${var.auto_default}-${var.cli}-${var.auto}-${var.config}"
}
