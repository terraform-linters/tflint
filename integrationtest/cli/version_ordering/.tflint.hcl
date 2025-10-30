plugin "foo" {
  enabled = true
}

plugin "bar" {
  enabled = true
  version = "0.1.0"
  source  = "github.com/terraform-linters/tflint-ruleset-bar"
}
