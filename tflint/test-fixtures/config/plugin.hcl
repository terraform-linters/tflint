config {
  disabled_by_default = true
}

rule "aws_instance_invalid_type" {
  enabled = false
}

rule "aws_instance_invalid_ami" {
  enabled = true
}

plugin "foo" {
  enabled = true

  custom = "foo"
}

plugin "bar" {
  enabled = false
}
