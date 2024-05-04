package cmd

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	flags "github.com/jessevdk/go-flags"
	"github.com/terraform-linters/tflint/terraform"
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
			Name:    "--call-module-type",
			Command: "./tflint --call-module-type all",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallAllModule,
				CallModuleTypeSet: true,
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
				CallModuleType:    terraform.CallLocalModule,
				Force:             true,
				ForceSet:          true,
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
				CallModuleType:    terraform.CallLocalModule,
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
			Name:    "multiple --ignore-module",
			Command: "./tflint --ignore-module module1 --ignore-module module2",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
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
				CallModuleType:    terraform.CallLocalModule,
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
			Name:    "multiple --var-file",
			Command: "./tflint --var-file example1.tfvars --var-file example2.tfvars",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
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
				CallModuleType:    terraform.CallLocalModule,
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
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules: map[string]*tflint.RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: true,
						Body:    nil,
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: true,
						Body:    nil,
					},
				},
				Plugins: map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "--disable-rule",
			Command: "./tflint --disable-rule aws_instance_invalid_type --disable-rule aws_instance_previous_type",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules: map[string]*tflint.RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: false,
						Body:    nil,
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: false,
						Body:    nil,
					},
				},
				Plugins: map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "--only",
			Command: "./tflint --only aws_instance_invalid_type",
			Expected: &tflint.Config{
				CallModuleType:       terraform.CallLocalModule,
				Force:                false,
				IgnoreModules:        map[string]bool{},
				Varfiles:             []string{},
				Variables:            []string{},
				DisabledByDefault:    true,
				DisabledByDefaultSet: true,
				Only:                 []string{"aws_instance_invalid_type"},
				Rules: map[string]*tflint.RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: true,
						Body:    nil,
					},
				},
				Plugins: map[string]*tflint.PluginConfig{},
			},
		},
		{
			Name:    "--enable-plugin",
			Command: "./tflint --enable-plugin test --enable-plugin another-test",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
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
		{
			Name:    "--format",
			Command: "./tflint --format compact",
			Expected: &tflint.Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Format:            "compact",
				FormatSet:         true,
				Rules:             map[string]*tflint.RuleConfig{},
				Plugins:           map[string]*tflint.PluginConfig{},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var opts Options
			parser := flags.NewParser(&opts, flags.HelpFlag)

			_, err := parser.ParseArgs(strings.Split(tc.Command, " "))
			if err != nil {
				t.Fatal(err)
			}

			ret := opts.toConfig()
			eqlopts := []cmp.Option{
				cmpopts.IgnoreUnexported(tflint.RuleConfig{}),
				cmpopts.IgnoreUnexported(tflint.PluginConfig{}),
				cmpopts.IgnoreUnexported(tflint.Config{}),
			}
			if diff := cmp.Diff(tc.Expected, ret, eqlopts...); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_toWorkerCommands(t *testing.T) {
	tests := []struct {
		name       string
		in         []string
		workingDir string
		want       []string
	}{
		{
			name:       "no args",
			in:         []string{},
			workingDir: "subdir",
			want:       []string{"--act-as-worker", "--chdir=subdir", "--force"},
		},
		{
			name: "all",
			in: []string{
				"--version",
				"--init",
				"--langserver",
				"--format=json",
				"--config=tflint.hcl",
				"--ignore-module=module1",
				"--ignore-module=module2",
				"--enable-rule=rule1",
				"--enable-rule=rule2",
				"--disable-rule=rule3",
				"--disable-rule=rule4",
				"--only=rule5",
				"--only=rule6",
				"--enable-plugin=plugin1",
				"--enable-plugin=plugin2",
				"--var-file=example1.tfvars",
				"--var-file=example2.tfvars",
				"--var=foo=bar",
				"--var=bar=baz",
				"--call-module-type=all",
				"--chdir=dir",
				"--recursive",
				"--filter=main1.tf",
				"--filter=main2.tf",
				"--force",
				"--minimum-failure-severity=warning",
				"--color",
				"--no-color",
				"--fix",
				"--no-parallel-runners",
				"--max-workers=2",
				"--act-as-bundled-plugin",
				"--act-as-worker",
			},
			workingDir: "subdir",
			want: []string{
				// "--version",
				// "--init",
				// "--langserver",
				// "--format=json",
				"--config=tflint.hcl",
				"--ignore-module=module1",
				"--ignore-module=module2",
				"--enable-rule=rule1",
				"--enable-rule=rule2",
				"--disable-rule=rule3",
				"--disable-rule=rule4",
				"--only=rule5",
				"--only=rule6",
				"--enable-plugin=plugin1",
				"--enable-plugin=plugin2",
				"--var-file=example1.tfvars",
				"--var-file=example2.tfvars",
				"--var=foo=bar",
				"--var=bar=baz",
				"--call-module-type=all",
				"--chdir=subdir", // "--chdir=dir",
				// "--recursive",
				"--filter=main1.tf",
				"--filter=main2.tf",
				"--force",
				// "--minimum-failure-severity=warning",
				// "--color",
				// "--no-color",
				"--fix",
				"--no-parallel-runners",
				// "--max-workers=2",
				// "--act-as-bundled-plugin",
				"--act-as-worker",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var in Options
			parser := flags.NewParser(&in, flags.HelpFlag)
			_, err := parser.ParseArgs(test.in)
			if err != nil {
				t.Fatal(err)
			}

			got := in.toWorkerCommands(test.workingDir)

			opt := cmpopts.SortSlices(func(a, b string) bool { return a < b })
			if diff := cmp.Diff(test.want, got, opt); diff != "" {
				t.Fatal(diff)
			}

			// Check if the output can be parsed
			var out Options
			parser = flags.NewParser(&out, flags.HelpFlag)
			_, err = parser.ParseArgs(got)
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
