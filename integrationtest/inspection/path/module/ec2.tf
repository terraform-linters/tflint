variable "root_suffix" {}
variable "module_suffix" {}

resource "aws_instance" "path_root" {
  ami = "ami-12345678"
  instance_type = "${path.root}/${var.root_suffix}"
}

resource "aws_instance" "path_module" {
  ami = "ami-12345678"
  instance_type = "${path.module}/${var.module_suffix}"
}
