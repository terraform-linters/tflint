package cmd

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	flags "github.com/jessevdk/go-flags"
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
				Force:             false,
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
				Force:             true,
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
				Force:             false,
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
				Force:             false,
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
				Force:             false,
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
				Force:             false,
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
				Force:             false,
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
				Force:             false,
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
				Force:             false,
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
				Force:             false,
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
		{
			Name:    "--enable-plugin",
			Command: "./tflint --enable-plugin test --enable-plugin another-test",
			Expected: &tflint.Config{
				Module:            false,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins: map[string]*tflint.PluginConfig{
					"test": {
						Name:    "test",
						Enabled: true,
						Body:    nil,
					},
					"another-test": {
						Name:    "another-test",
						Enabled: true,
						Body:    nil,
					},
				},
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
		eqlopts := []cmp.Option{
			cmpopts.IgnoreUnexported(tflint.RuleConfig{}),
			cmpopts.IgnoreUnexported(tflint.PluginConfig{}),
			cmpopts.IgnoreUnexported(tflint.Config{}),
		}
		if !cmp.Equal(tc.Expected, ret, eqlopts...) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(tc.Expected, ret, eqlopts...))
		}
	}
}
