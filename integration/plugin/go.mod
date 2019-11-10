module github.com/terraform-linters/tflint/plugin/test-fixtures/plugins

go 1.13

require (
	github.com/hashicorp/hcl/v2 v2.0.0
	github.com/terraform-linters/tflint v0.0.0
)

replace github.com/terraform-linters/tflint v0.0.0 => ../../
