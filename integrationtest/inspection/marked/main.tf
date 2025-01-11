variable "no_marked" {
  default   = "t2.micro"
}

variable "sensitive" {
  sensitive = true
  default   = "t2.micro"
}

variable "ephemeral" {
  ephemeral = true
  default   = "t2.micro"
}

variable "marked_set" {
  sensitive = true
  default   = [true]
}

resource "aws_instance" "no_marked" {
  instance_type = var.no_marked
}

resource "aws_instance" "sensitive" {
  instance_type = var.sensitive
}

resource "aws_instance" "ephemeral" {
  instance_type = var.ephemeral
}

resource "aws_s3_bucket" "main" {
  dynamic "lifecycle_rule" {
    for_each = var.marked_set

    content {
      enabled = lifecycle_rule.value
    }
  }

  dynamic "lifecycle_rule" {
    for_each = var.marked_set

    content {
      enabled = true
    }
  }
}
