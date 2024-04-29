resource "aws_instance" "count" {
  count = 2

  instance_type = "t${count.index}.micro"
}

resource "aws_instance" "for_each" {
  for_each = {
    v1 = "micro"
    v2 = "medium"
  }

  instance_type = "${each.key}.${each.value}"
}

module "count" {
  source = "./module"
  count  = 2

  instance_type = "t${count.index}.micro"
}

module "for_each" {
  source = "./module"
  for_each = {
    v1 = "micro"
    v2 = "medium"
  }

  instance_type = "${each.key}.${each.value}"
}

variable "sensitive" {
  sensitive = true
}

resource "aws_instance" "sensitive" {
  count = 1

  instance_type = "${count.index}.${var.sensitive}"
}

resource "aws_instance" "tags" {
  count = 1

  tags = {
    count = count.index
    sensitive = var.sensitive
  }
}

resource "aws_instance" "provider_function" {
  count = 1

  instance_type = "${count.index}.${provider::tflint::instance_type()}"
}
