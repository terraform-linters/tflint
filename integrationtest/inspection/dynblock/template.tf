resource "aws_s3_bucket" "bucket" {
  lifecycle_rule {
    enabled = false
    transition {
      days          = 30
      storage_class = "STANDARD_IA"
    }
  }
}

resource "aws_s3_bucket" "dynamic" {
  dynamic "cors_rule" {
    for_each = toset(["*"])

    content {
      allowed_headers = [cors_rule.value]
    }
  }

  dynamic "lifecycle_rule" {
    for_each = toset([true, false])

    content {
      enabled = var.force_disable ? false : lifecycle_rule.value

      dynamic "transition" {
        for_each = toset([30, 60])

        content {
          days          = lifecycle_rule.value ? transition.value + 10 : transition.value
          storage_class = "STANDARD_IA"
        }
      }
    }
  }
}

variable "force_disable" {
  type    = bool
  default = false
}

resource "aws_s3_bucket" "dynamic_with_meta_arguments" {
  for_each = toset([true, false])

  dynamic "lifecycle_rule" {
    for_each = toset([true, false])

    content {
      enabled = lifecycle_rule.value && each.value
    }
  }
}
