module github.com/terraform-linters/tflint

go 1.14

require (
	github.com/aws/aws-sdk-go v1.32.5
	github.com/fatih/color v1.9.0
	github.com/golang/mock v1.4.3
	github.com/google/go-cmp v0.4.1
	github.com/hashicorp/aws-sdk-go-base v0.4.0
	github.com/hashicorp/go-plugin v1.3.0
	github.com/hashicorp/go-version v1.2.0
	github.com/hashicorp/hcl/v2 v2.6.0
	github.com/hashicorp/logutils v1.0.0
	github.com/hashicorp/terraform v0.12.26
	github.com/hashicorp/terraform-plugin-sdk v1.14.0 // indirect
	github.com/jessevdk/go-flags v1.4.0
	github.com/mattn/go-colorable v0.1.6
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sourcegraph/go-lsp v0.0.0-20181119182933-0c7d621186c1
	github.com/sourcegraph/jsonrpc2 v0.0.0-20190106185902-35a74f039c6a
	github.com/spf13/afero v1.2.2
	github.com/terraform-linters/tflint-plugin-sdk v0.1.2-0.20200615160547-c1d3caf80fe0
	github.com/terraform-providers/terraform-provider-aws v2.65.0+incompatible // indirect
	github.com/zclconf/go-cty v1.5.0
)

replace github.com/terraform-providers/terraform-provider-aws v2.65.0+incompatible => github.com/terraform-providers/terraform-provider-aws v1.60.1-0.20200604234259-3853d337c01a
