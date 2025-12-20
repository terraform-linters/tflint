resource "aws_prometheus_workspace" "main" {
  lifecycle {
    ignore_changes = [
      logging_configuration
    ]
  }
}
