package tflint

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	hcl "github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/afero"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	sdk "github.com/terraform-linters/tflint-plugin-sdk/tflint"
)

func TestLoadConfig(t *testing.T) {
	// default error check helper
	neverHappend := func(err error) bool { return err != nil }

	tests := []struct {
		name     string
		file     string
		files    map[string]string
		want     *Config
		errCheck func(error) bool
	}{
		{
			name: "load file",
			file: "config.hcl",
			files: map[string]string{
				"config.hcl": `
config {
	plugin_dir = "~/.tflint.d/plugins"

	module = true
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
				Module: true,
				Force:  true,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-module": true,
				},
				Varfiles:          []string{"example1.tfvars", "example2.tfvars"},
				Variables:         []string{"foo=bar", "bar=['foo']"},
				DisabledByDefault: false,
				PluginDir:         "~/.tflint.d/plugins",
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
						SourceOwner: "foo",
						SourceRepo:  "bar",
					},
					"baz": {
						Name:    "baz",
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
			want:     EmptyConfig(),
			errCheck: neverHappend,
		},
		{
			name: "default home config",
			file: ".tflint.hcl",
			files: map[string]string{
				"/root/.tflint.hcl": `
config {
	force = true
	disabled_by_default = true
}`,
			},
			want: &Config{
				Module:            false,
				Force:             true,
				IgnoreModules:     map[string]bool{},
				Varfiles:          []string{},
				Variables:         []string{},
				DisabledByDefault: true,
				Rules:             map[string]*RuleConfig{},
				Plugins:           map[string]*PluginConfig{},
			},
			errCheck: neverHappend,
		},
		{
			name:     "no config",
			file:     ".tflint.hcl",
			want:     EmptyConfig(),
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
				return err == nil || err.Error() != "plugin `foo`: `source` attribute cannot be omitted when specifying `version`"
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
				return err == nil || err.Error() != "plugin `foo`: `version` attribute cannot be omitted when specifying `source`"
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
				return err == nil || err.Error() != "plugin `foo`: `source` is invalid. Must be in the format `github.com/owner/repo`"
			},
		},
		{
			name: "plugin with invalid source host",
			file: "plugin_with_invalid_source_host.hcl",
			files: map[string]string{
				"plugin_with_invalid_source_host.hcl": `
plugin "foo" {
	enabled = true

	version = "0.1.0"
	source = "gitlab.com/foo/bar"
}`,
			},
			errCheck: func(err error) bool {
				return err == nil || err.Error() != "plugin `foo`: `source` is invalid. Hostname must be `github.com`"
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("HOME", "/root")
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
		Module: true,
		Force:  true,
		IgnoreModules: map[string]bool{
			"github.com/terraform-linters/example-1": true,
			"github.com/terraform-linters/example-2": false,
		},
		Varfiles:          []string{"example1.tfvars", "example2.tfvars"},
		Variables:         []string{"foo=bar"},
		DisabledByDefault: false,
		PluginDir:         "./.tflint.d/plugins",
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
				Module: true,
				Force:  false,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-1": true,
					"github.com/terraform-linters/example-2": false,
				},
				Varfiles:          []string{"example1.tfvars", "example2.tfvars"},
				Variables:         []string{"foo=bar"},
				DisabledByDefault: false,
				PluginDir:         "./.tflint.d/plugins",
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
				Module: false,
				Force:  true,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-2": true,
					"github.com/terraform-linters/example-3": false,
				},
				Varfiles:          []string{"example3.tfvars"},
				Variables:         []string{"bar=baz"},
				DisabledByDefault: false,
				PluginDir:         "~/.tflint.d/plugins",
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
				Module: true,
				Force:  true,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-1": true,
					"github.com/terraform-linters/example-2": true,
					"github.com/terraform-linters/example-3": false,
				},
				Varfiles:          []string{"example1.tfvars", "example2.tfvars", "example3.tfvars"},
				Variables:         []string{"foo=bar", "bar=baz"},
				DisabledByDefault: false,
				PluginDir:         "~/.tflint.d/plugins",
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
				Module: true,
				Force:  false,
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
				Module: false,
				Force:  true,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-2": true,
					"github.com/terraform-linters/example-3": false,
				},
				Varfiles:          []string{"example3.tfvars"},
				Variables:         []string{"bar=baz"},
				DisabledByDefault: true,
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
				Module: true,
				Force:  true,
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-1": true,
					"github.com/terraform-linters/example-2": true,
					"github.com/terraform-linters/example-3": false,
				},
				Varfiles:          []string{"example1.tfvars", "example2.tfvars", "example3.tfvars"},
				Variables:         []string{"foo=bar", "bar=baz"},
				DisabledByDefault: true,
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
				Module:            false,
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
				Module:            false,
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
				Module:            false,
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
				Module:            false,
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
				Module:            false,
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
				Module:            false,
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
			plugin, exists := config.Plugins["test"]
			if !exists {
				t.Fatal("plugin `test` should be declared")
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
			Err:      errors.New("`aws_instance_invalid_ami` is duplicated in ruleSetB and ruleSetB"),
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
