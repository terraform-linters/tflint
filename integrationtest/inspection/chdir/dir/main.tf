variable "from_config" {
  default = "default"
}

variable "from_cli" {
  default = "default"
}

variable "from_auto" {
  default = "default"
}

variable "from_auto_default" {
  default = "default"
}

module "aws_instance" {
  source = "./module"

  instance_type = "${var.from_config}-${var.from_cli}-${var.from_auto}-${var.from_auto_default}-${file("dir.txt")}-${file("${path.cwd}/root.txt")}"
}
