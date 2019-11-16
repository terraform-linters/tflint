package tflint

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/terraform-linters/tflint/client"
)

func Test_LoadConfig(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name     string
		File     string
		Fallback string
		Expected *Config
	}{
		{
			Name: "load file",
			File: filepath.Join(currentDir, "test-fixtures", "config", "config.hcl"),
			Expected: &Config{
				Module:    true,
				DeepCheck: true,
				Force:     true,
				AwsCredentials: client.AwsCredentials{
					AccessKey: "AWS_ACCESS_KEY",
					SecretKey: "AWS_SECRET_KEY",
					Region:    "us-east-1",
					Profile:   "production",
					CredsFile: "~/.aws/myapp",
				},
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-module": true,
				},
				Varfiles:  []string{"example1.tfvars", "example2.tfvars"},
				Variables: []string{"foo=bar", "bar=['foo']"},
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
			},
		},
		{
			Name:     "empty file",
			File:     filepath.Join(currentDir, "test-fixtures", "config", "empty.hcl"),
			Expected: EmptyConfig(),
		},
		{
			Name:     "fallback",
			File:     filepath.Join(currentDir, "test-fixtures", "config", "not_found.hcl"),
			Fallback: filepath.Join(currentDir, "test-fixtures", "config", "fallback.hcl"),
			Expected: &Config{
				Module:    false,
				DeepCheck: true,
				Force:     true,
				AwsCredentials: client.AwsCredentials{
					AccessKey: "AWS_ACCESS_KEY",
					SecretKey: "AWS_SECRET_KEY",
					Region:    "us-east-1",
				},
				IgnoreModules: map[string]bool{},
				Varfiles:      []string{},
				Variables:     []string{},
				Rules:         map[string]*RuleConfig{},
			},
		},
		{
			Name:     "fallback file not found",
			File:     filepath.Join(currentDir, "test-fixtures", "config", "not_found.hcl"),
			Fallback: filepath.Join(currentDir, "test-fixtures", "config", "fallback_not_found.hcl"),
			Expected: EmptyConfig(),
		},
	}

	for _, tc := range cases {
		originalDefault := defaultConfigFile
		defaultConfigFile = filepath.Join(currentDir, "test-fixtures", "config", "not_found.hcl")
		originalFallback := fallbackConfigFile
		fallbackConfigFile = tc.Fallback

		ret, err := LoadConfig(tc.File)
		if err != nil {
			t.Fatalf("Failed `%s` test: Unexpected error occurred: %s", tc.Name, err)
		}

		if !cmp.Equal(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(tc.Expected, ret))
		}

		defaultConfigFile = originalDefault
		fallbackConfigFile = originalFallback
	}
}

func Test_LoadConfig_error(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name     string
		File     string
		Expected string
	}{
		{
			Name: "file not found",
			File: filepath.Join(currentDir, "test-fixtures", "config", "not_found.hcl"),
			Expected: fmt.Sprintf(
				"`%s` is not found",
				filepath.Join(currentDir, "test-fixtures", "config", "not_found.hcl"),
			),
		},
		{
			Name: "syntax error",
			File: filepath.Join(currentDir, "test-fixtures", "config", "syntax_error.hcl"),
			Expected: fmt.Sprintf(
				"%s:1,1-2: Invalid character; The \"`\" character is not valid. To create a multi-line string, use the \"heredoc\" syntax, like \"<<EOT\"., and 1 other diagnostic(s)",
				filepath.Join(currentDir, "test-fixtures", "config", "syntax_error.hcl"),
			),
		},
		{
			Name: "invalid config",
			File: filepath.Join(currentDir, "test-fixtures", "config", "invalid.hcl"),
			Expected: fmt.Sprintf(
				"%s:1,34-42: Extraneous label for rule; Only 1 labels (name) are expected for rule blocks.",
				filepath.Join(currentDir, "test-fixtures", "config", "invalid.hcl"),
			),
		},
		{
			Name:     "terraform_version",
			File:     filepath.Join(currentDir, "test-fixtures", "config", "terraform_version.hcl"),
			Expected: "`terraform_version` was removed in v0.9.0 because the option is no longer used",
		},
		{
			Name:     "ignore_rule",
			File:     filepath.Join(currentDir, "test-fixtures", "config", "ignore_rule.hcl"),
			Expected: "`ignore_rule` was removed in v0.12.0. Please define `rule` block with `enabled = false` instead",
		},
	}

	for _, tc := range cases {
		_, err := LoadConfig(tc.File)
		if err == nil {
			t.Fatalf("Failed `%s` test: Expected error does not occurred", tc.Name)
		}

		if err.Error() != tc.Expected {
			t.Fatalf("Failed `%s` test: expected error is `%s`, but get `%s`", tc.Name, tc.Expected, err.Error())
		}
	}
}

func Test_Merge(t *testing.T) {
	cfg := &Config{
		Module:    true,
		DeepCheck: true,
		Force:     true,
		AwsCredentials: client.AwsCredentials{
			AccessKey: "access_key",
			SecretKey: "secret_key",
			Region:    "us-east-1",
		},
		IgnoreModules: map[string]bool{
			"github.com/terraform-linters/example-1": true,
			"github.com/terraform-linters/example-2": false,
		},
		Varfiles:  []string{"example1.tfvars", "example2.tfvars"},
		Variables: []string{"foo=bar"},
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
		Base     *Config
		Other    *Config
		Expected *Config
	}{
		{
			Name:     "empty",
			Base:     EmptyConfig(),
			Other:    EmptyConfig(),
			Expected: EmptyConfig(),
		},
		{
			Name:     "prefer base",
			Base:     cfg,
			Other:    EmptyConfig(),
			Expected: cfg,
		},
		{
			Name:     "prefer other",
			Base:     EmptyConfig(),
			Other:    cfg,
			Expected: cfg,
		},
		{
			Name: "override and merge",
			Base: &Config{
				Module:    true,
				DeepCheck: true,
				Force:     false,
				AwsCredentials: client.AwsCredentials{
					AccessKey: "access_key",
					SecretKey: "secret_key",
					Profile:   "production",
					Region:    "us-east-1",
				},
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-1": true,
					"github.com/terraform-linters/example-2": false,
				},
				Varfiles:  []string{"example1.tfvars", "example2.tfvars"},
				Variables: []string{"foo=bar"},
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
			},
			Other: &Config{
				Module:    false,
				DeepCheck: false,
				Force:     true,
				AwsCredentials: client.AwsCredentials{
					AccessKey: "ACCESS_KEY",
					SecretKey: "SECRET_KEY",
					Region:    "ap-northeast-1",
					CredsFile: "~/.aws/myapp",
				},
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-2": true,
					"github.com/terraform-linters/example-3": false,
				},
				Varfiles:  []string{"example3.tfvars"},
				Variables: []string{"bar=baz"},
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_ami": {
						Name:    "aws_instance_invalid_ami",
						Enabled: false,
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: false,
					},
				},
			},
			Expected: &Config{
				Module:    true,
				DeepCheck: true, // DeepCheck will not override
				Force:     true,
				AwsCredentials: client.AwsCredentials{
					AccessKey: "ACCESS_KEY",
					SecretKey: "SECRET_KEY",
					Profile:   "production",
					Region:    "ap-northeast-1",
					CredsFile: "~/.aws/myapp",
				},
				IgnoreModules: map[string]bool{
					"github.com/terraform-linters/example-1": true,
					"github.com/terraform-linters/example-2": true,
					"github.com/terraform-linters/example-3": false,
				},
				Varfiles:  []string{"example1.tfvars", "example2.tfvars", "example3.tfvars"},
				Variables: []string{"foo=bar", "bar=baz"},
				Rules: map[string]*RuleConfig{
					"aws_instance_invalid_type": {
						Name:    "aws_instance_invalid_type",
						Enabled: false,
					},
					"aws_instance_invalid_ami": {
						Name:    "aws_instance_invalid_ami",
						Enabled: false,
					},
					"aws_instance_previous_type": {
						Name:    "aws_instance_previous_type",
						Enabled: false,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		ret := tc.Base.Merge(tc.Other)
		if !cmp.Equal(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test: diff=%s", tc.Name, cmp.Diff(tc.Expected, ret))
		}
	}
}

func Test_copy(t *testing.T) {
	cfg := &Config{
		Module:    true,
		DeepCheck: true,
		Force:     true,
		AwsCredentials: client.AwsCredentials{
			AccessKey: "access_key",
			SecretKey: "secret_key",
			Region:    "us-east-1",
		},
		IgnoreModules: map[string]bool{
			"github.com/terraform-linters/example-1": true,
			"github.com/terraform-linters/example-2": false,
		},
		Varfiles:  []string{"example1.tfvars", "example2.tfvars"},
		Variables: []string{},
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
		Name       string
		SideEffect func(*Config)
	}{
		{
			Name: "Module",
			SideEffect: func(c *Config) {
				c.Module = false
			},
		},
		{
			Name: "DeepCheck",
			SideEffect: func(c *Config) {
				c.DeepCheck = false
			},
		},
		{
			Name: "Force",
			SideEffect: func(c *Config) {
				c.Force = false
			},
		},
		{
			Name: "AwsCredentials",
			SideEffect: func(c *Config) {
				c.AwsCredentials = client.AwsCredentials{
					Profile: "production",
					Region:  "us-east-1",
				}
			},
		},
		{
			Name: "IgnoreModules",
			SideEffect: func(c *Config) {
				c.IgnoreModules["github.com/terraform-linters/example-1"] = false
			},
		},
		{
			Name: "Varfiles",
			SideEffect: func(c *Config) {
				c.Varfiles = append(c.Varfiles, "new.tfvars")
			},
		},
		{
			Name: "Variables",
			SideEffect: func(c *Config) {
				c.Variables = append(c.Variables, "baz=foo")
			},
		},
		{
			Name: "Rules",
			SideEffect: func(c *Config) {
				c.Rules["aws_instance_invalid_type"].Enabled = true
			},
		},
	}

	for _, tc := range cases {
		ret := cfg.copy()
		if !cmp.Equal(cfg, ret) {
			t.Fatalf("The copied config doesn't match original: Diff=%s", cmp.Diff(cfg, ret))
		}

		tc.SideEffect(ret)
		if cmp.Equal(cfg, ret) {
			t.Fatalf("The original was changed when updating `%s`", tc.Name)
		}
	}
}
