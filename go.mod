module github.com/terraform-linters/tflint

go 1.15

require (
	github.com/aws/aws-sdk-go v1.34.9
	github.com/fatih/color v1.9.0
	github.com/golang/mock v1.4.4
	github.com/google/go-cmp v0.5.2
	github.com/hashicorp/aws-sdk-go-base v0.6.0
	github.com/hashicorp/go-plugin v1.3.0
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/hashicorp/logutils v1.0.0
	github.com/hashicorp/terraform v0.13.1
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.0.1
	github.com/jessevdk/go-flags v1.4.0
	github.com/jstemmer/go-junit-report v0.9.1
	github.com/mattn/go-colorable v0.1.7
	github.com/mitchellh/go-homedir v1.1.0
	github.com/serenize/snaker v0.0.0-20171204205717-a683aaf2d516
	github.com/sourcegraph/go-lsp v0.0.0-20181119182933-0c7d621186c1
	github.com/sourcegraph/jsonrpc2 v0.0.0-20190106185902-35a74f039c6a
	github.com/spf13/afero v1.3.4
	github.com/terraform-linters/tflint-plugin-sdk v0.4.1-0.20200822151013-70ed6c361b0b
	github.com/terraform-providers/terraform-provider-aws v3.3.0+incompatible
	github.com/zclconf/go-cty v1.5.1
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
)

replace github.com/terraform-providers/terraform-provider-aws v3.3.0+incompatible => github.com/terraform-providers/terraform-provider-aws v1.60.1-0.20200820211857-51f8bae0d4ee
