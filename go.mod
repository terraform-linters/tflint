module github.com/terraform-linters/tflint

go 1.16

require (
	github.com/fatih/color v1.10.0
	github.com/golang/mock v1.5.0
	github.com/google/go-cmp v0.5.5
	github.com/hashicorp/go-plugin v1.4.0
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/hcl/v2 v2.9.1
	github.com/hashicorp/logutils v1.0.0
	github.com/hashicorp/terraform v0.14.8
	github.com/jessevdk/go-flags v1.4.0
	github.com/jstemmer/go-junit-report v0.9.1
	github.com/mattn/go-colorable v0.1.8
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sourcegraph/go-lsp v0.0.0-20181119182933-0c7d621186c1
	github.com/sourcegraph/jsonrpc2 v0.0.0-20190106185902-35a74f039c6a
	github.com/spf13/afero v1.2.2 // matches version used by terraform
	github.com/terraform-linters/tflint-plugin-sdk v0.8.2-0.20210320093036-ae6749e72de0
	github.com/terraform-linters/tflint-ruleset-aws v0.3.1-0.20210320093704-d770eccc81e0
	github.com/zclconf/go-cty v1.8.1
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
)
