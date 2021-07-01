module github.com/terraform-linters/tflint

go 1.16

require (
	github.com/Masterminds/semver/v3 v3.1.1
	github.com/agext/levenshtein v1.2.3
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/apparentlymart/go-versions v1.0.1
	github.com/bmatcuk/doublestar v1.3.4
	github.com/fatih/color v1.12.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/go-cmp v0.5.6
	github.com/google/go-github/v35 v35.3.0
	github.com/google/uuid v1.2.0
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/hashicorp/errwrap v1.1.0
	github.com/hashicorp/go-getter v1.5.5
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-plugin v1.4.1
	github.com/hashicorp/go-uuid v1.0.2
	github.com/hashicorp/go-version v1.3.0
	github.com/hashicorp/hcl/v2 v2.10.0
	github.com/hashicorp/logutils v1.0.0
	github.com/hashicorp/terraform-svchost v0.0.0-20200729002733-f050f53b9734
	github.com/jessevdk/go-flags v1.5.0
	github.com/jstemmer/go-junit-report v0.9.1
	github.com/mattn/go-colorable v0.1.8
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sourcegraph/go-lsp v0.0.0-20181119182933-0c7d621186c1
	github.com/sourcegraph/jsonrpc2 v0.0.0-20190106185902-35a74f039c6a
	github.com/spf13/afero v1.2.2 // matches version used by terraform
	github.com/terraform-linters/tflint-plugin-sdk v0.8.3-0.20210614125323-8364139f3745
	github.com/terraform-linters/tflint-ruleset-aws v0.4.1
	github.com/zclconf/go-cty v1.8.4
	github.com/zclconf/go-cty-yaml v1.0.2
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/mod v0.4.2
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	golang.org/x/term v0.0.0-20201210144234-2321bbc49cbf // indirect
	golang.org/x/text v0.3.6
	google.golang.org/api v0.34.0 // indirect
)
