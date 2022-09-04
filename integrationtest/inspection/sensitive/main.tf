variable "sensitive" {
  sensitive = true
  default   = "t2.micro"
}

variable "non_sensitive" {
  sensitive = false
  default   = "t2.micro"
}

resource "aws_instance" "sensitive" {
  instance_type = var.sensitive
}

resource "aws_instance" "non_sensitive" {
  instance_type = var.non_sensitive
}
