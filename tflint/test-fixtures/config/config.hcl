config {
  plugin_dir = "~/.tflint.d/plugins"

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
  version = "0.1.0"
  source = "github.com/foo/bar"
  signing_key = "SIGNING_KEY"
}
