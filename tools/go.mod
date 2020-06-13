module github.com/terraform-linters/tflint/tools

go 1.14

require (
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/hashicorp/terraform-plugin-sdk v1.13.1
	github.com/serenize/snaker v0.0.0-20171204205717-a683aaf2d516
	github.com/terraform-linters/tflint-plugin-sdk v0.1.2-0.20200613135440-8f8f107dab08
	github.com/terraform-providers/terraform-provider-aws v2.65.0+incompatible
)

replace github.com/terraform-providers/terraform-provider-aws v2.65.0+incompatible => github.com/terraform-providers/terraform-provider-aws v1.60.1-0.20200604234259-3853d337c01a
