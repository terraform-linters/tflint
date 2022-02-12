variable "annotation" {
  default = "activate/beta1"
}

provider "custom" {
  zone = "asia"
  annotation {
    value = var.annotation
  }
}

resource "aws_instance" "foo" {
  instance_type = "t2.micro"
}
