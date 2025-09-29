package tflint

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint/terraform"
)

func TestLoadConfig(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		name     string
		file     string
		files    map[string]string
		envs     map[string]string
		want     *Config
		errCheck func(error) bool
	}{
		{
			name: "load file",
			file: "config.hcl",
			files: map[string]string{
				"config.hcl": `
tflint {
	required_version = ">= 0"
}

config {
	format = "compact"
	plugin_dir = "~/.tflint.d/plugins"

	call_module_type = "all"
	force = true

	ignore_module = {
		"github.com/terraform-linters/example-module" = true
	}

	varfile = ["example1.tfvars", "example2.tfvars"]

	variables = ["foo=bar", "bar=['foo']"]
}

rule "aws_instance_invalid_type" {
	enabled = false
}

rule "aws_instance_previous_type" {
	enabled = false
	foo = "bar"
}

plugin "foo" {
	enabled = true
}

plugin "bar" {
	enabled = false
	version = "0.1.0"
	source = "github.com/foo/bar"
	signing_key = "SIGNING_KEY"
}

plugin "baz" {
	enabled = true
	foo = "baz"
}`,
			},
			want: &Config{
				CallModuleType:    terraform.CallAllModule,
				CallModuleTypeSet: true,
				Force:             true,
				ForceSet:          true,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-module": true,
				},
				Varfiles:          []string{"example1.tfvars", "example2.tfvars"},
				Variables:         []string{"foo=bar", "bar=['foo']"},
				DisabledByDefault: false,
				PluginDir:         "~/.tflint.d/plugins",
				PluginDirSet:      true,
				Format:            "compact",
				FormatSet:         true,
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: false,
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: false,
					},
				},
				Plugins: map[string]*PluginConfig{
					"foo": {
						Name:    "foo",
						Enabled: true,
					},
					"bar": {
						Name:        "bar",
						Enabled:     false,
						Version:     "0.1.0",
						Source:      "github.com/foo/bar",
						SigningKey:  "SIGNING_KEY",
						SourceHost:  "github.com",
						SourceOwner: "foo",
						SourceRepo:  "bar",
					},
					"baz": {
						Name:    "baz",
						Enabled: true,
					},
					"terraform": {
						Name:    "terraform",
						Enabled: true,
					},
				},
			},
			errCheck: neverHappend,
		},
		{
			name:     "empty file",
			file:     "empty.hcl",
			files:    map[string]string{"empty.hcl": ""},
			want:     EmptyConfig().enableBundledPlugin(),
			errCheck: neverHappend,
		},
		{
			name: "TFLINT_CONFIG_FILE",
			file: "",
			files: map[string]string{
				"env.hcl": `
config {
	force = true
	disabled_by_default = true
}`,
			},
			envs: map[string]string{
				"TFLINT_CONFIG_FILE": "env.hcl",
			},
			want: &Config{
				CallModuleType:       terraform.CallLocalModule,
				Force:                true,
				ForceSet:             true,
				IgnoreModules:        map[string]bool{},
				Varfiles:             []string{},
				Variables:            []string{},
				DisabledByDefault:    true,
				DisabledByDefaultSet: true,
				Rules:                map[string]*RuleConfig{},
				Plugins: map[string]*PluginConfig{
					"terraform": {
						Name:    "terraform",
						Enabled: true,
					},
				},
			},
			errCheck: neverHappend,
		},
		{
			name: "default home config",
			file: "",
			files: map[string]string{
				"/root/.tflint.hcl": `
config {
	force = true
	disabled_by_default = true
}`,
			},
			want: &Config{
				CallModuleType:       terraform.CallLocalModule,
				Force:                true,
				ForceSet:             true,
				IgnoreModules:        map[string]bool{},
				Varfiles:             []string{},
				Variables:            []string{},
				DisabledByDefault:    true,
				DisabledByDefaultSet: true,
				Rules:                map[string]*RuleConfig{},
				Plugins: map[string]*PluginConfig{
					"terraform": {
						Name:    "terraform",
						Enabled: true,
					},
				},
			},
			errCheck: neverHappend,
		},
		{
			name:     "no config",
			file:     "",
			want:     EmptyConfig().enableBundledPlugin(),
			errCheck: neverHappend,
		},
		{
			name: "terraform plugin",
			file: "config.hcl",
			files: map[string]string{
				"config.hcl": `
plugin "terraform" {
  enabled = false
}`,
			},
			want: &Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*RuleConfig{},
				Plugins: map[string]*PluginConfig{
					"terraform": {
						Name:    "terraform",
						Enabled: false,
					},
				},
			},
			errCheck: neverHappend,
		},
		{
			name: "file not found",
			file: "not_found.hcl",
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "failed to load file: open not_found.hcl: file does not exist"
			},
		},
		{
			name: "file not found with TFLINT_CONFIG_FILE",
			file: "",
			envs: map[string]string{
				"TFLINT_CONFIG_FILE": "not_found.hcl",
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "failed to load file: open not_found.hcl: file does not exist"
			},
		},
		{
			name: "syntax error",
			file: "syntax_error.hcl",
			files: map[string]string{
				"syntax_error.hcl": `}`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "syntax_error.hcl:1,1-2: Argument or block definition required; An argument or block definition is required here."
			},
		},
		{
			name: "invalid config",
			file: "invalid.hcl",
			files: map[string]string{
				"invalid.hcl": `
rule "aws_instance_invalid_type" "myrule" {
	enabled = false
}`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "invalid.hcl:2,34-42: Extraneous label for rule; Only 1 labels (name) are expected for rule blocks."
			},
		},
		{
			name: "invalid format",
			file: "invalid_format.hcl",
			files: map[string]string{
				"invalid_format.hcl": `
config {
	format = "invalid"
}`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "invalid is invalid format. Allowed formats are: default, json, checkstyle, junit, compact, sarif"
			},
		},
		{
			name: "invalid call_module_type",
			file: "invalid_call_module_type.hcl",
			files: map[string]string{
				"invalid_call_module_type.hcl": `
config {
	call_module_type = "invalid"
}`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "invalid is invalid call module type. Allowed values are: all, local, none"
			},
		},
		{
			name: "plugin without source",
			file: "plugin_without_source.hcl",
			files: map[string]string{
				"plugin_without_source.hcl": `
plugin "foo" {
	enabled = true

	version = "0.1.0"
}`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `plugin "foo": "source" attribute cannot be omitted when specifying "version"`
			},
		},
		{
			name: "plugin without version",
			file: "plugin_without_version.hcl",
			files: map[string]string{
				"plugin_without_version.hcl": `
plugin "foo" {
	enabled = true

	source = "github.com/foo/bar"
}`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `plugin "foo": "version" attribute cannot be omitted when specifying "source"`
			},
		},
		{
			name: "plugin with invalid source",
			file: "plugin_with_invalid_source.hcl",
			files: map[string]string{
				"plugin_with_invalid_source.hcl": `
plugin "foo" {
	enabled = true

	version = "0.1.0"
	source = "github.com/foo/bar/baz"
}`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `plugin "foo": "source" is invalid. Must be a GitHub reference in the format "${host}/${owner}/${repo}"`
			},
		},
		{
			name: "plugin with GHES source host",
			file: "plugin_with_ghes_source_host.hcl",
			files: map[string]string{
				"plugin_with_ghes_source_host.hcl": `
plugin "foo" {
	enabled = true

	version = "0.1.0"
	source = "github.example.com/foo/bar"
}`,
			},
			want: &Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*RuleConfig{},
				Plugins: map[string]*PluginConfig{
					"foo": {
						Name:        "foo",
						Enabled:     true,
						Version:     "0.1.0",
						Source:      "github.example.com/foo/bar",
						SourceHost:  "github.example.com",
						SourceOwner: "foo",
						SourceRepo:  "bar",
					},
					"terraform": {
						Name:    "terraform",
						Enabled: true,
					},
				},
			},
			errCheck: neverHappend,
		},
		{
			name: "prefer the passed file over TFLINT_CONFIG_FILE",
			file: "cli.hcl",
			files: map[string]string{
				"cli.hcl": `
config {
	force = true
	disabled_by_default = false
}`,
				"env.hcl": `
config {
	force = false
	disabled_by_default = true
}`,
			},
			envs: map[string]string{
				"TFLINT_CONFIG_FILE": "env.hcl",
			},
			want: &Config{
				CallModuleType:       terraform.CallLocalModule,
				Force:                true,
				ForceSet:             true,
				IgnoreModules:        map[string]bool{},
				Varfiles:             []string{},
				Variables:            []string{},
				DisabledByDefault:    false,
				DisabledByDefaultSet: true,
				Rules:                map[string]*RuleConfig{},
				Plugins: map[string]*PluginConfig{
					"terraform": {
						Name:    "terraform",
						Enabled: true,
					},
				},
			},
			errCheck: neverHappend,
		},
		{
			name: "valid required_version",
			file: "config.hcl",
			files: map[string]string{
				"config.hcl": `
tflint {
  required_version = ">= 0.50"
}`,
			},
			want:     EmptyConfig().enableBundledPlugin(),
			errCheck: neverHappend,
		},
		{
			name: "invalid required_version",
			file: "config.hcl",
			files: map[string]string{
				"config.hcl": `
tflint {
  required_version = "< 0.50"
}`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != fmt.Sprintf(`config.hcl:3,22-30: Unsupported TFLint version; This config does not support the currently used TFLint version %s. Please update to another supported version or change the version requirement.`, Version)
			},
		},
		{
			name: "multiple required_versions",
			file: "config.hcl",
			files: map[string]string{
				"config.hcl": `
tflint {
  required_version = ">= 0.50"
}

tflint {
  required_version = "< 0.50"
}`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `config.hcl:6,1-7: Multiple "tflint" blocks are not allowed; The "tflint" block is already found in config.hcl:2,1-7, but found the second one.`
			},
		},
		{
			name: "removed module attribute",
			file: "config.hcl",
			files: map[string]string{
				"config.hcl": `
config {
  module = true
}`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != `"module" attribute was removed in v0.54.0. Use "call_module_type" instead`
			},
		},
		{
			name: "load JSON config file",
			file: "config.json",
			files: map[string]string{
				"config.json": `{
  "tflint": {
    "required_version": ">= 0"
  },
  "config": {
    "format": "compact",
    "plugin_dir": "~/.tflint.d/plugins",
    "call_module_type": "all",
    "force": true,
    "ignore_module": {
      "github.com/terraform-linters/example-module": true
    },
    "varfile": ["example1.tfvars", "example2.tfvars"],
    "variables": ["foo=bar", "bar=['foo']"]
  },
  "rule": {
    "aws_instance_invalid_type": {
      "enabled": false
    },
    "aws_instance_previous_type": {
      "enabled": false,
      "foo": "bar"
    }
  },
  "plugin": {
    "foo": {
      "enabled": true
    },
    "bar": {
      "enabled": false,
      "version": "0.1.0",
      "source": "github.com/foo/bar",
      "signing_key": "SIGNING_KEY"
    },
    "baz": {
      "enabled": true,
      "foo": "baz"
    }
  }
}`,
			},
			want: &Config{
				CallModuleType:    terraform.CallAllModule,
				CallModuleTypeSet: true,
				Force:             true,
				ForceSet:          true,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-module": true,
				},
				Varfiles:          []string{"example1.tfvars", "example2.tfvars"},
				Variables:         []string{"foo=bar", "bar=['foo']"},
				DisabledByDefault: false,
				PluginDir:         "~/.tflint.d/plugins",
				PluginDirSet:      true,
				Format:            "compact",
				FormatSet:         true,
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: false,
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: false,
					},
				},
				Plugins: map[string]*PluginConfig{
					"foo": {
						Name:    "foo",
						Enabled: true,
					},
					"bar": {
						Name:        "bar",
						Enabled:     false,
						Version:     "0.1.0",
						Source:      "github.com/foo/bar",
						SigningKey:  "SIGNING_KEY",
						SourceHost:  "github.com",
						SourceOwner: "foo",
						SourceRepo:  "bar",
					},
					"baz": {
						Name:    "baz",
						Enabled: true,
					},
					"terraform": {
						Name:    "terraform",
						Enabled: true,
					},
				},
			},
			errCheck: neverHappend,
		},
		{
			name: "empty JSON file",
			file: "empty.json",
			files: map[string]string{
				"empty.json": "{}",
			},
			want:     EmptyConfig().enableBundledPlugin(),
			errCheck: neverHappend,
		},
		{
			name: "JSON syntax error",
			file: "syntax_error.json",
			files: map[string]string{
				"syntax_error.json": `{"config": {`,
			},
			errCheck: func(err error) bool {
				return err == nil || !strings.Contains(err.Error(), "syntax_error.json")
			},
		},
		{
			name: "default JSON config file",
			file: "",
			files: map[string]string{
				".tflint.json": `{
  "config": {
    "force": true,
    "disabled_by_default": true
  }
}`,
			},
			want: &Config{
				CallModuleType:       terraform.CallLocalModule,
				Force:                true,
				ForceSet:             true,
				IgnoreModules:        map[string]bool{},
				Varfiles:             []string{},
				Variables:            []string{},
				DisabledByDefault:    true,
				DisabledByDefaultSet: true,
				Rules:                map[string]*RuleConfig{},
				Plugins: map[string]*PluginConfig{
					"terraform": {
						Name:    "terraform",
						Enabled: true,
					},
				},
			},
			errCheck: neverHappend,
		},
		{
			name: "prefer HCL over JSON in current directory",
			file: "",
			files: map[string]string{
				".tflint.hcl":  `config { force = false }`,
				".tflint.json": `{"config": {"force": true}}`,
			},
			want: &Config{
				CallModuleType: terraform.CallLocalModule,
				Force:          false,
				ForceSet:       true,
				IgnoreModules:  map[string]bool{},
				Varfiles:       []string{},
				Variables:      []string{},
				Rules:          map[string]*RuleConfig{},
				Plugins: map[string]*PluginConfig{
					"terraform": {
						Name:    "terraform",
						Enabled: true,
					},
				},
			},
			errCheck: neverHappend,
		},
		{
			name: "home directory JSON config",
			file: "",
			files: map[string]string{
				"/root/.tflint.json": `{
  "config": {
    "force": true,
    "disabled_by_default": true
  }
}`,
			},
			want: &Config{
				CallModuleType:       terraform.CallLocalModule,
				Force:                true,
				ForceSet:             true,
				IgnoreModules:        map[string]bool{},
				Varfiles:             []string{},
				Variables:            []string{},
				DisabledByDefault:    true,
				DisabledByDefaultSet: true,
				Rules:                map[string]*RuleConfig{},
				Plugins: map[string]*PluginConfig{
					"terraform": {
						Name:    "terraform",
						Enabled: true,
					},
				},
			},
			errCheck: neverHappend,
		},
		{
			name: "TFLINT_CONFIG_FILE with JSON",
			file: "",
			files: map[string]string{
				"env.json": `{
  "config": {
    "force": true,
    "disabled_by_default": true
  }
}`,
			},
			envs: map[string]string{
				"TFLINT_CONFIG_FILE": "env.json",
			},
			want: &Config{
				CallModuleType:       terraform.CallLocalModule,
				Force:                true,
				ForceSet:             true,
				IgnoreModules:        map[string]bool{},
				Varfiles:             []string{},
				Variables:            []string{},
				DisabledByDefault:    true,
				DisabledByDefaultSet: true,
				Rules:                map[string]*RuleConfig{},
				Plugins: map[string]*PluginConfig{
					"terraform": {
						Name:    "terraform",
						Enabled: true,
					},
				},
			},
			errCheck: neverHappend,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("HOME", "/root")
			for k, v := range test.envs {
				t.Setenv(k, v)
			}
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			for name, src := range test.files {
				if err := fs.WriteFile(name, []byte(src), os.ModePerm); err != nil {
					t.Fatal(err)
				}
			}

			got, err := LoadConfig(fs, test.file)
			if test.errCheck(err) {
				t.Fatal(err)
			}

			opts := []cmp.Option{
				cmpopts.IgnoreUnexported(Config{}),
				cmpopts.IgnoreFields(PluginConfig{}, "Body"),
				cmpopts.IgnoreFields(RuleConfig{}, "Body"),
			}
			if diff := cmp.Diff(test.want, got, opts...); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestMerge(t *testing.T) {
	file1, diags := hclsyntax.ParseConfig([]byte(`foo = "bar"`), "test.hcl", hcl.Pos{})
	if diags.HasErrors() {
		t.Fatalf("Failed to parse test config: %s", diags)
	}
	file2, diags := hclsyntax.ParseConfig([]byte(`bar = "baz"`), "test2.hcl", hcl.Pos{})
	if diags.HasErrors() {
		t.Fatalf("Failed to parse test config: %s", diags)
	}

	config := &Config{
		CallModuleType:    terraform.CallAllModule,
		CallModuleTypeSet: true,
		Force:             true,
		ForceSet:          true,
		IgnoreModules: map[string]bool{
			"github.com/terraform-linters/example-1": true,
			"github.com/terraform-linters/example-2": false,
		},
		Varfiles:          []string{"example1.tfvars", "example2.tfvars"},
		Variables:         []string{"foo=bar"},
		DisabledByDefault: false,
		PluginDir:         "./.tflint.d/plugins",
		PluginDirSet:      true,
		Format:            "compact",
		FormatSet:         true,
		Rules: map[string]*RuleConfig{
			"aws_instance_invalid_type": {
				Name:    "aws_instance_invalid_type",
				Enabled: false,
				Body:    file1.Body,
			},
			"aws_instance_invalid_ami": {
				Name:    "aws_instance_invalid_ami",
				Enabled: true,
				Body:    file2.Body,
			},
		},
		Plugins: map[string]*PluginConfig{},
	}

	tests := []struct {
		name  string
		base  *Config
		other *Config
		want  *Config
	}{
		{
			name:  "empty",
			base:  EmptyConfig(),
			other: EmptyConfig(),
			want:  EmptyConfig(),
		},
		{
			name:  "prefer base",
			base:  config,
			other: EmptyConfig(),
			want:  config,
		},
		{
			name:  "prefer other",
			base:  EmptyConfig(),
			other: config,
			want:  config,
		},
		{
			name: "override and merge",
			base: &Config{
				CallModuleType:    terraform.CallNoModule,
				CallModuleTypeSet: true,
				Force:             false,
				ForceSet:          true,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-1": true,
					"github.com/terraform-linters/example-2": false,
				},
				Varfiles:             []string{"example1.tfvars", "example2.tfvars"},
				Variables:            []string{"foo=bar"},
				DisabledByDefault:    true,
				DisabledByDefaultSet: true,
				PluginDir:            "./.tflint.d/plugins",
				PluginDirSet:         true,
				Format:               "compact",
				FormatSet:            true,
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: false,
						Body:    file1.Body,
					},
					"aws_instance_invalid_ami": {
						Name:    "aws_instance_invalid_ami",
						Enabled: true,
						Body:    file1.Body,
					},
				},
				Plugins: map[string]*PluginConfig{
					"foo": {
						Name:    "foo",
						Enabled: true,
					},
					"bar": {
						Name:    "bar",
						Enabled: false,
					},
				},
			},
			other: &Config{
				CallModuleType:    terraform.CallAllModule,
				CallModuleTypeSet: true,
				Force:             true,
				ForceSet:          true,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-2": true,
					"github.com/terraform-linters/example-3": false,
				},
				Varfiles:             []string{"example3.tfvars"},
				Variables:            []string{"bar=baz"},
				DisabledByDefault:    false,
				DisabledByDefaultSet: true,
				PluginDir:            "~/.tflint.d/plugins",
				PluginDirSet:         true,
				Format:               "json",
				FormatSet:            true,
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_ami": {
						Name:    "aws_instance_invalid_ami",
						Enabled: false,
						Body:    file2.Body,
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: false,
						Body:    file2.Body,
					},
				},
				Plugins: map[string]*PluginConfig{
					"baz": {
						Name:    "baz",
						Enabled: true,
					},
					"bar": {
						Name:    "bar",
						Enabled: true,
					},
				},
			},
			want: &Config{
				CallModuleType:    terraform.CallAllModule,
				CallModuleTypeSet: true,
				Force:             true,
				ForceSet:          true,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-1": true,
					"github.com/terraform-linters/example-2": true,
					"github.com/terraform-linters/example-3": false,
				},
				Varfiles:             []string{"example1.tfvars", "example2.tfvars", "example3.tfvars"},
				Variables:            []string{"foo=bar", "bar=baz"},
				DisabledByDefault:    false,
				DisabledByDefaultSet: true,
				PluginDir:            "~/.tflint.d/plugins",
				PluginDirSet:         true,
				Format:               "json",
				FormatSet:            true,
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: false,
						Body:    file1.Body,
					},
					"aws_instance_invalid_ami": {
						Name:    "aws_instance_invalid_ami",
						Enabled: false,
						Body:    file2.Body,
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: false,
						Body:    file2.Body,
					},
				},
				Plugins: map[string]*PluginConfig{
					"foo": {
						Name:    "foo",
						Enabled: true,
					},
					"bar": {
						Name:    "bar",
						Enabled: true,
					},
					"baz": {
						Name:    "baz",
						Enabled: true,
					},
				},
			},
		},
		{
			name: "CLI --only argument and merge",
			base: &Config{
				CallModuleType:    terraform.CallAllModule,
				CallModuleTypeSet: true,
				Force:             false,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-1": true,
					"github.com/terraform-linters/example-2": false,
				},
				Varfiles:          []string{"example1.tfvars", "example2.tfvars"},
				Variables:         []string{"foo=bar"},
				DisabledByDefault: false,
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: false,
						Body:    file1.Body,
					},
					"aws_instance_invalid_ami": {
						Name:    "aws_instance_invalid_ami",
						Enabled: false,
						Body:    file1.Body,
					},
				},
				Plugins: map[string]*PluginConfig{
					"foo": {
						Name:    "foo",
						Enabled: true,
					},
					"bar": {
						Name:    "bar",
						Enabled: false,
					},
				},
			},
			other: &Config{
				CallModuleType: terraform.CallLocalModule,
				Force:          true,
				ForceSet:       true,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-2": true,
					"github.com/terraform-linters/example-3": false,
				},
				Varfiles:             []string{"example3.tfvars"},
				Variables:            []string{"bar=baz"},
				DisabledByDefault:    true,
				DisabledByDefaultSet: true,
				Only:                 []string{"aws_instance_invalid_type", "aws_instance_previous_type"},
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: true,
						Body:    nil,
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: true,
						Body:    nil, // Will not attach, missing config
					},
				},
				Plugins: map[string]*PluginConfig{
					"baz": {
						Name:    "baz",
						Enabled: true,
					},
					"bar": {
						Name:    "bar",
						Enabled: true,
					},
				},
			},
			want: &Config{
				CallModuleType:    terraform.CallAllModule,
				CallModuleTypeSet: true,
				Force:             true,
				ForceSet:          true,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-1": true,
					"github.com/terraform-linters/example-2": true,
					"github.com/terraform-linters/example-3": false,
				},
				Varfiles:             []string{"example1.tfvars", "example2.tfvars", "example3.tfvars"},
				Variables:            []string{"foo=bar", "bar=baz"},
				DisabledByDefault:    true,
				DisabledByDefaultSet: true,
				Only:                 []string{"aws_instance_invalid_type", "aws_instance_previous_type"},
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: true, // Passing an --only rule from CLI will always enable
						Body:    file1.Body,
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: true,
						Body:    nil,
					},
					"aws_instance_invalid_ami": {
						Name:    "aws_instance_invalid_ami",
						Enabled: false,
						Body:    file1.Body,
					},
				},
				Plugins: map[string]*PluginConfig{
					"foo": {
						Name:    "foo",
						Enabled: true,
					},
					"bar": {
						Name:    "bar",
						Enabled: true,
					},
					"baz": {
						Name:    "baz",
						Enabled: true,
					},
				},
			},
		},
		{
			name: "merge rule config with CLI-based config",
			base: &Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: false,
						Body:    file1.Body,
					},
				},
				Plugins: map[string]*PluginConfig{},
			},
			other: &Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: true,
						Body:    nil,
					},
				},
				Plugins: map[string]*PluginConfig{},
			},
			want: &Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: true,       // overridden
						Body:    file1.Body, // keep
					},
				},
				Plugins: map[string]*PluginConfig{},
			},
		},
		{
			name: "merge plugin config with CLI-based config",
			base: &Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*RuleConfig{},
				Plugins: map[string]*PluginConfig{
					"aws": {
						Name:        "aws",
						Enabled:     false,
						Version:     "0.1.0",
						Source:      "github.com/terraform-linters/tflint-ruleset-aws",
						SigningKey:  "key",
						Body:        file1.Body,
						SourceOwner: "terraform-linters",
						SourceRepo:  "tflint-ruleset-aws",
					},
				},
			},
			other: &Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*RuleConfig{},
				Plugins: map[string]*PluginConfig{
					"aws": {
						Name:    "aws",
						Enabled: true,
					},
				},
			},
			want: &Config{
				CallModuleType:    terraform.CallLocalModule,
				Force:             false,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: false,
				Rules:             map[string]*RuleConfig{},
				Plugins: map[string]*PluginConfig{
					"aws": {
						Name:        "aws",
						Enabled:     true, // overridden
						Version:     "0.1.0",
						Source:      "github.com/terraform-linters/tflint-ruleset-aws",
						SigningKey:  "key",
						Body:        file1.Body,
						SourceOwner: "terraform-linters",
						SourceRepo:  "tflint-ruleset-aws",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.base.Merge(test.other)

			opts := []cmp.Option{
				cmpopts.IgnoreUnexported(Config{}),
				cmpopts.IgnoreUnexported(hclsyntax.Body{}),
				cmpopts.IgnoreFields(hclsyntax.Body{}, "Attributes", "Blocks"),
			}
			if diff := cmp.Diff(test.want, test.base, opts...); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func Test_ToPluginConfig(t *testing.T) {
	src := `
config {
	disabled_by_default = true
}

rule "aws_instance_invalid_type" {
	enabled = false
}

rule "aws_instance_invalid_ami" {
	enabled = true
}

plugin "foo" {
	enabled = true

	custom = "foo"
}

plugin "bar" {
	enabled = false
}`

	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	if err := fs.WriteFile("test.hcl", []byte(src), os.ModePerm); err != nil {
		t.Fatal(err)
	}
	config, err := LoadConfig(fs, "test.hcl")
	if err != nil {
		t.Fatal(err)
	}
	config.Only = []string{"aws_instance_invalid_ami"}

	got := config.ToPluginConfig()
	want := &sdk.Config{
		Rules: map[string]*sdk.RuleConfig{
			"aws_instance_invalid_type": {
				Name:    "aws_instance_invalid_type",
				Enabled: false,
			},
			"aws_instance_invalid_ami": {
				Name:    "aws_instance_invalid_ami",
				Enabled: true,
			},
		},
		DisabledByDefault: true,
		Only:              []string{"aws_instance_invalid_ami"},
	}
	opts := cmp.Options{
		cmpopts.IgnoreUnexported(PluginConfig{}),
		cmpopts.IgnoreFields(hcl.Range{}, "Filename"),
		cmpopts.IgnoreFields(hcl.Pos{}, "Byte"),
	}
	if diff := cmp.Diff(got, want, opts...); diff != "" {
		t.Fatal(diff)
	}
}

func TestPluginContent(t *testing.T) {
	tests := []struct {
		Name      string
		Config    string
		CLI       bool
		Arg       *hclext.BodySchema
		Want      *hclext.BodyContent
		DiagCount int
	}{
		{
			Name: "schema is nil",
			Config: `
plugin "test" {
	enabled = true
}`,
			Arg: nil,
			Want: &hclext.BodyContent{
				Attributes: hclext.Attributes{},
				Blocks:     hclext.Blocks{},
			},
			DiagCount: 0,
		},
		{
			Name: "manually installed plugin",
			Config: `
plugin "test" {
	enabled = true
	foo = "bar"
}`,
			Arg: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
			},
			Want: &hclext.BodyContent{
				Attributes: hclext.Attributes{
					"foo": &hclext.Attribute{Name: "foo"},
				},
				Blocks: hclext.Blocks{},
			},
			DiagCount: 0,
		},
		{
			Name: "auto installed plugin",
			Config: `
plugin "test" {
	enabled = true
	version = "0.1.0"
	source  = "github.com/example/example"

	signing_key = "PUBLIC_KEY"

	foo = "bar"
}`,
			Arg: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
			},
			Want: &hclext.BodyContent{
				Attributes: hclext.Attributes{
					"foo": &hclext.Attribute{Name: "foo"},
				},
				Blocks: hclext.Blocks{},
			},
			DiagCount: 0,
		},
		{
			Name: "enabled by CLI",
			CLI:  true,
			Arg: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
			},
			Want:      &hclext.BodyContent{},
			DiagCount: 0,
		},
		{
			Name: "required attribute is not found",
			Config: `
plugin "test" {
	enabled = true
}`,
			Arg: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo", Required: true}},
			},
			DiagCount: 1, // The argument "foo" is required
		},
		{
			Name: "unsupported attribute is found",
			Config: `
plugin "test" {
	enabled = true
	bar = "baz"
}`,
			Arg: &hclext.BodySchema{
				Attributes: []hclext.AttributeSchema{{Name: "foo"}},
			},
			DiagCount: 1, // An argument named "bar" is not expected here
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			if err := fs.WriteFile("test.hcl", []byte(test.Config), os.ModePerm); err != nil {
				t.Fatal(err)
			}
			config, err := LoadConfig(fs, "test.hcl")
			if err != nil {
				t.Fatal(err)
			}
			var plugin *PluginConfig
			if test.CLI {
				plugin = &PluginConfig{Name: "test", Body: nil}
			} else {
				var exists bool
				plugin, exists = config.Plugins["test"]
				if !exists {
					t.Fatal(`plugin "test" should be declared`)
				}
			}

			content, diags := plugin.Content(test.Arg)
			if len(diags) != test.DiagCount {
				t.Errorf("wrong number of diagnostics %d; want %d", len(diags), test.DiagCount)
				for _, diag := range diags {
					t.Logf(" - %s", diag.Error())
				}
			}
			if diags.HasErrors() {
				return
			}

			opts := cmp.Options{
				cmpopts.IgnoreFields(hclext.Attribute{}, "Expr", "Range", "NameRange"),
			}
			if diff := cmp.Diff(content, test.Want, opts); diff != "" {
				t.Errorf("diff: %s", diff)
			}
		})
	}
}

type ruleSetA struct{}

func (*ruleSetA) RuleSetName() (string, error) {
	return "ruleSetA", nil
}
func (*ruleSetA) RuleSetVersion() (string, error) {
	return "0.1.0", nil
}
func (*ruleSetA) RuleNames() ([]string, error) {
	return []string{"aws_instance_invalid_type"}, nil
}

type ruleSetB struct{}

func (*ruleSetB) RuleSetName() (string, error) {
	return "ruleSetB", nil
}
func (*ruleSetB) RuleSetVersion() (string, error) {
	return "0.1.0", nil
}
func (*ruleSetB) RuleNames() ([]string, error) {
	return []string{"aws_instance_invalid_ami"}, nil
}

func Test_ValidateRules(t *testing.T) {
	config := &Config{
		Rules: map[string]*RuleConfig{
			"aws_instance_invalid_type": {
				Name:    "aws_instance_invalid_type",
				Enabled: false,
			},
			"aws_instance_invalid_ami": {
				Name:    "aws_instance_invalid_ami",
				Enabled: true,
			},
		},
	}

	cases := []struct {
		Name     string
		Config   *Config
		RuleSets []RuleSet
		Err      error
	}{
		{
			Name:     "valid",
			Config:   config,
			RuleSets: []RuleSet{&ruleSetA{}, &ruleSetB{}},
			Err:      nil,
		},
		{
			Name:     "duplicate",
			Config:   config,
			RuleSets: []RuleSet{&ruleSetA{}, &ruleSetB{}, &ruleSetB{}},
			Err:      errors.New(`"aws_instance_invalid_ami" is duplicated in ruleSetB and ruleSetB`),
		},
		{
			Name:     "not found",
			Config:   config,
			RuleSets: []RuleSet{&ruleSetB{}},
			Err:      errors.New("Rule not found: aws_instance_invalid_type"),
		},
	}

	for _, tc := range cases {
		err := tc.Config.ValidateRules(tc.RuleSets...)

		if tc.Err == nil {
			if err != nil {
				t.Fatalf("Failed %s: Unexpected error occurred: %s", tc.Name, err)
			}
		} else {
			if err == nil {
				t.Fatalf("Failed %s: Error should have occurred, but didn't", tc.Name)
			}
			if err.Error() != tc.Err.Error() {
				t.Fatalf("Failed %s: error message is not matched: want=%s got=%s", tc.Name, tc.Err, err)
			}
		}
	}
}

func TestConfigSources(t *testing.T) {
	tests := []struct {
		name        string
		file        string
		files       map[string]string
		wantSources map[string]string // filename -> expected content
		errCheck    func(error) bool
	}{
		{
			name: "HCL config sources preserved",
			file: "config.hcl",
			files: map[string]string{
				"config.hcl": `config {
  format = "compact"
}

plugin "terraform" {
  enabled = true
  preset = "all"
}`,
			},
			wantSources: map[string]string{
				"config.hcl": `config {
  format = "compact"
}

plugin "terraform" {
  enabled = true
  preset = "all"
}`,
				bundledPluginConfigFilename: bundledPluginConfigContent,
			},
			errCheck: func(err error) bool { return err != nil },
		},
		{
			name: "JSON config sources preserved",
			file: "config.json",
			files: map[string]string{
				"config.json": `{
  "config": {
    "format": "json"
  },
  "plugin": {
    "terraform": {
      "enabled": true,
      "preset": "all"
    }
  }
}`,
			},
			wantSources: map[string]string{
				"config.json": `{
  "config": {
    "format": "json"
  },
  "plugin": {
    "terraform": {
      "enabled": true,
      "preset": "all"
    }
  }
}`,
				bundledPluginConfigFilename: bundledPluginConfigContent,
			},
			errCheck: func(err error) bool { return err != nil },
		},
		{
			name: "default JSON config sources",
			file: "",
			files: map[string]string{
				".tflint.json": `{
  "config": {
    "force": true
  }
}`,
			},
			wantSources: map[string]string{
				".tflint.json": `{
  "config": {
    "force": true
  }
}`,
				bundledPluginConfigFilename: bundledPluginConfigContent,
			},
			errCheck: func(err error) bool { return err != nil },
		},
		{
			name: "mixed HCL bundled + JSON user config",
			file: "user.json",
			files: map[string]string{
				"user.json": `{
  "rule": {
    "terraform_unused_declarations": {
      "enabled": false
    }
  }
}`,
			},
			wantSources: map[string]string{
				"user.json": `{
  "rule": {
    "terraform_unused_declarations": {
      "enabled": false
    }
  }
}`,
				bundledPluginConfigFilename: bundledPluginConfigContent,
			},
			errCheck: func(err error) bool { return err != nil },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fs := afero.Afero{Fs: afero.NewMemMapFs()}
			for name, src := range test.files {
				if err := fs.WriteFile(name, []byte(src), os.ModePerm); err != nil {
					t.Fatal(err)
				}
			}

			config, err := LoadConfig(fs, test.file)
			if test.errCheck(err) {
				t.Fatal(err)
			}

			sources := config.Sources()

			// Verify all expected sources are present
			for filename, expectedContent := range test.wantSources {
				actualContent, exists := sources[filename]
				if !exists {
					t.Errorf("Expected source file %q not found in sources map", filename)
					continue
				}

				if string(actualContent) != expectedContent {
					t.Errorf("Source content mismatch for %q:\nwant: %s\ngot:  %s", filename, expectedContent, string(actualContent))
				}
			}

			// Verify no unexpected sources are present
			for filename := range sources {
				if _, expected := test.wantSources[filename]; !expected {
					t.Errorf("Unexpected source file %q found in sources map", filename)
				}
			}
		})
	}
}
