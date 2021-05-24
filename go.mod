module github.com/terraform-linters/tflint

go 1.16

require (
	github.com/fatih/color v1.10.0
	github.com/golang/mock v1.5.0
	github.com/google/go-cmp v0.5.5
	github.com/google/go-github/v35 v35.2.0
	github.com/hashicorp/go-plugin v1.4.1
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/hcl/v2 v2.10.0
	github.com/hashicorp/logutils v1.0.0
	github.com/hashicorp/terraform v0.15.3
	github.com/jessevdk/go-flags v1.5.0
	github.com/jstemmer/go-junit-report v0.9.1
	github.com/mattn/go-colorable v0.1.8
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sourcegraph/go-lsp v0.0.0-20181119182933-0c7d621186c1
	github.com/sourcegraph/jsonrpc2 v0.0.0-20190106185902-35a74f039c6a
	github.com/spf13/afero v1.2.2 // matches version used by terraform
	github.com/terraform-linters/tflint-plugin-sdk v0.8.2
	github.com/terraform-linters/tflint-ruleset-aws v0.4.0
	github.com/zclconf/go-cty v1.8.3
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
)
