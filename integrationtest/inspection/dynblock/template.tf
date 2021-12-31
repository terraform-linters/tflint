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
  dynamic "lifecycle_rule" {
    for_each = toset([true])

    content {
      enabled = lifecycle_rule.value

      dynamic "transition" {
        for_each = toset([30])

        content {
          days          = transition.value
          storage_class = "STANDARD_IA"
        }
      }
    }
  }
}
