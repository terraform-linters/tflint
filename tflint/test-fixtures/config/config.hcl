config {
  module = true
  deep_check = true
  force = true

  aws_credentials = {
    access_key = "AWS_ACCESS_KEY"
    secret_key = "AWS_SECRET_KEY"
    region     = "us-east-1"
  }

  ignore_rule = {
    aws_instance_invalid_type  = true
    aws_instance_previous_type = true
  }

  ignore_module = {
    "github.com/wata727/example-module" = true
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
