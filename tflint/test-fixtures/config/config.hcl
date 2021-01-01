config {
  module = true
  force = true

  ignore_module = {
    "github.com/terraform-linters/example-module" = true
  }

  varfile = ["example1.tfvars", "example2.tfvars"]

  variables = ["foo=bar", "bar=['foo']"]
}

rule "aws_instance_invalid_type" {
  enabled = false
}

rule "aws_instance_previous_type" {
  enabled = false
}

plugin "foo" {
  enabled = true
}

plugin "bar" {
  enabled = false
}
