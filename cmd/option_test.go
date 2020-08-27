package cmd

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	hcl "github.com/hashicorp/hcl/v2"
	flags "github.com/jessevdk/go-flags"
	"github.com/terraform-linters/tflint/client"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_toConfig(t *testing.T) {
	cases := []struct {
		Name     string
		Command  string
		Expected *tflint.Config
	}{
		{
			Name:     "default",
			Command:  "./tflint",
			Expected: tflint.EmptyConfig(),
		},
		{
			Name:    "--module",
			Command: "./tflint --module",
			Expected: &tflint.Config{
				Module:            true,
				DeepCheck:         false,
				Force:             false,
				AwsCredentials:    client.AwsCredentials{},
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins:           map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "--deep",
			Command: "./tflint --deep",
			Expected: &tflint.Config{
				Module:            false,
				DeepCheck:         true,
				Force:             false,
				AwsCredentials:    client.AwsCredentials{},
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins:           map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "--force",
			Command: "./tflint --force",
			Expected: &tflint.Config{
				Module:            false,
				DeepCheck:         false,
				Force:             true,
				AwsCredentials:    client.AwsCredentials{},
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins:           map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "AWS static credentials",
			Command: "./tflint --aws-access-key AWS_ACCESS_KEY_ID --aws-secret-key AWS_SECRET_ACCESS_KEY --aws-region us-east-1",
			Expected: &tflint.Config{
				Module:    false,
				DeepCheck: false,
				Force:     false,
				AwsCredentials: client.AwsCredentials{
					AccessKey: "AWS_ACCESS_KEY_ID",
					SecretKey: "AWS_SECRET_ACCESS_KEY",
					Region:    "us-east-1",
				},
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins:           map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "AWS shared credentials",
			Command: "./tflint --aws-profile production --aws-region us-east-1",
			Expected: &tflint.Config{
				Module:    false,
				DeepCheck: false,
				Force:     false,
				AwsCredentials: client.AwsCredentials{
					Profile: "production",
					Region:  "us-east-1",
				},
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins:           map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "AWS shared credentials in another file",
			Command: "./tflint --aws-creds-file ~/.aws/myapp --aws-profile production --aws-region us-east-1",
			Expected: &tflint.Config{
				Module:    false,
				DeepCheck: false,
				Force:     false,
				AwsCredentials: client.AwsCredentials{
					CredsFile: "~/.aws/myapp",
					Profile:   "production",
					Region:    "us-east-1",
				},
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins:           map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "--ignore-module",
			Command: "./tflint --ignore-module module1,module2",
			Expected: &tflint.Config{
				Module:            false,
				DeepCheck:         false,
				Force:             false,
				AwsCredentials:    client.AwsCredentials{},
				IgnoreModules:     map[string]bool{"module1": true, "module2": true},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins:           map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "multiple `--ignore-module`",
			Command: "./tflint --ignore-module module1 --ignore-module module2",
			Expected: &tflint.Config{
				Module:            false,
				DeepCheck:         false,
				Force:             false,
				AwsCredentials:    client.AwsCredentials{},
				IgnoreModules:     map[string]bool{"module1": true, "module2": true},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins:           map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "--var-file",
			Command: "./tflint --var-file example1.tfvars,example2.tfvars",
			Expected: &tflint.Config{
				Module:            false,
				DeepCheck:         false,
				Force:             false,
				AwsCredentials:    client.AwsCredentials{},
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{"example1.tfvars", "example2.tfvars"},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins:           map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "multiple `--var-file`",
			Command: "./tflint --var-file example1.tfvars --var-file example2.tfvars",
			Expected: &tflint.Config{
				Module:            false,
				DeepCheck:         false,
				Force:             false,
				AwsCredentials:    client.AwsCredentials{},
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{"example1.tfvars", "example2.tfvars"},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins:           map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "--var",
			Command: "./tflint --var foo=bar --var bar=baz",
			Expected: &tflint.Config{
				Module:            false,
				DeepCheck:         false,
				Force:             false,
				AwsCredentials:    client.AwsCredentials{},
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{"foo=bar", "bar=baz"},
				DisabledByDefault: false,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins:           map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "--enable-rule",
			Command: "./tflint --enable-rule aws_instance_invalid_type --enable-rule aws_instance_previous_type",
			Expected: &tflint.Config{
				Module:            false,
				DeepCheck:         false,
				Force:             false,
				AwsCredentials:    client.AwsCredentials{},
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules: map[string]*tflint.RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: true,
						Body:    hcl.EmptyBody(),
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: true,
						Body:    hcl.EmptyBody(),
					},
				},
				Plugins: map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "--disable-rule",
			Command: "./tflint --disable-rule aws_instance_invalid_type --disable-rule aws_instance_previous_type",
			Expected: &tflint.Config{
				Module:            false,
				DeepCheck:         false,
				Force:             false,
				AwsCredentials:    client.AwsCredentials{},
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules: map[string]*tflint.RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: false,
						Body:    hcl.EmptyBody(),
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: false,
						Body:    hcl.EmptyBody(),
					},
				},
				Plugins: map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "--only",
			Command: "./tflint --only aws_instance_invalid_type",
			Expected: &tflint.Config{
				Module:            false,
				DeepCheck:         false,
				Force:             false,
				AwsCredentials:    client.AwsCredentials{},
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: true,
				Rules: map[string]*tflint.RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: true,
						Body:    hcl.EmptyBody(),
					},
				},
				Plugins: map[string]*tflint.PluginConfig{},
			},
		},
	}

	for _, tc := range cases {
		var opts Options
		parser := flags.NewParser(&opts, flags.HelpFlag)

		_, err := parser.ParseArgs(strings.Split(tc.Command, " "))
		if err != nil {
			t.Fatalf("Failed `%s` test: %s", tc.Name, err)
		}

		ret := opts.toConfig()
		if !cmp.Equal(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(tc.Expected, ret))
		}
	}
}
