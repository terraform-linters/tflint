variable "unknown_set" {
  type = set(bool)
}

variable "unknown_bool" {
  type = bool
}

resource "aws_s3_bucket" "main" {
  lifecycle_rule {
    enabled = true
  }

  dynamic "lifecycle_rule" {
    for_each = var.unknown_set

    content {
      enabled = lifecycle_rule.value
    }
  }

  dynamic "lifecycle_rule" {
    for_each = toset([var.unknown_bool])

    content {
      enabled = lifecycle_rule.value
    }
  }
}

resource "aws_iam_role" "main" {
  inline_policy {
    name = "static"
  }

  dynamic "inline_policy" {
    for_each = toset(["foo", "bar"])

    content {
      name = inline_policy.value
    }
  }

  dynamic "inline_policy" {
    for_each = var.unknown_set

    content {
      name = inline_policy.value
    }
  }
}

resource "testing_assertions" "main" {
  equal "static" {}

  dynamic "equal" {
    for_each = toset(["known_label"])
    iterator = it
    labels   = [it.value]
    content {}
  }

  dynamic "equal" {
    for_each = toset(["unknown_label"])
    iterator = it
    labels   = ["${it.value}-${var.unknown_bool}"]
    content {}
  }
}
