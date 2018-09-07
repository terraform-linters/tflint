package tflint

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/wata727/tflint/client"
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
				DeepCheck: true,
				AwsCredentials: client.AwsCredentials{
					AccessKey: "AWS_ACCESS_KEY",
					SecretKey: "AWS_SECRET_KEY",
					Region:    "us-east-1",
				},
				IgnoreRule: map[string]bool{
					"aws_instance_invalid_type":  true,
					"aws_instance_previous_type": true,
				},
				IgnoreModule: map[string]bool{
					"github.com/wata727/example-module": true,
				},
				Varfile:          []string{"example1.tfvars", "example2.tfvars"},
				TerraformVersion: "0.9.11",
				Rules: map[string]*Rule{
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
				DeepCheck: true,
				AwsCredentials: client.AwsCredentials{
					AccessKey: "AWS_ACCESS_KEY",
					SecretKey: "AWS_SECRET_KEY",
					Region:    "us-east-1",
				},
				IgnoreRule:       map[string]bool{},
				IgnoreModule:     map[string]bool{},
				Varfile:          []string{},
				TerraformVersion: "0.9.11",
				Rules:            map[string]*Rule{},
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
				"%s:2,1-1: Invalid character; The \"`\" character is not valid. To create a multi-line string, use the \"heredoc\" syntax, like \"<<EOT\"., and 1 other diagnostic(s)",
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
		DeepCheck: true,
		AwsCredentials: client.AwsCredentials{
			AccessKey: "access_key",
			SecretKey: "secret_key",
			Region:    "us-east-1",
		},
		IgnoreModule: map[string]bool{
			"github.com/wata727/example-1": true,
			"github.com/wata727/example-2": false,
		},
		IgnoreRule: map[string]bool{
			"aws_instance_invalid_type": false,
			"aws_instance_invalid_ami":  true,
		},
		Varfile:          []string{"example1.tfvars", "example2.tfvars"},
		TerraformVersion: "0.11.1",
		Rules: map[string]*Rule{
			"aws_instance_invalid_type": &Rule{
				Name:    "aws_instance_invalid_type",
				Enabled: false,
			},
			"aws_instance_invalid_ami": &Rule{
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
				DeepCheck: true,
				AwsCredentials: client.AwsCredentials{
					AccessKey: "access_key",
					SecretKey: "secret_key",
					Profile:   "production",
					Region:    "us-east-1",
				},
				IgnoreModule: map[string]bool{
					"github.com/wata727/example-1": true,
					"github.com/wata727/example-2": false,
				},
				IgnoreRule: map[string]bool{
					"aws_instance_invalid_type": false,
					"aws_instance_invalid_ami":  true,
				},
				Varfile:          []string{"example1.tfvars", "example2.tfvars"},
				TerraformVersion: "0.11.1",
				Rules: map[string]*Rule{
					"aws_instance_invalid_type": &Rule{
						Name:    "aws_instance_invalid_type",
						Enabled: false,
					},
					"aws_instance_invalid_ami": &Rule{
						Name:    "aws_instance_invalid_ami",
						Enabled: true,
					},
				},
			},
			Other: &Config{
				DeepCheck: false,
				AwsCredentials: client.AwsCredentials{
					AccessKey: "ACCESS_KEY",
					SecretKey: "SECRET_KEY",
					Region:    "ap-northeast-1",
				},
				IgnoreModule: map[string]bool{
					"github.com/wata727/example-2": true,
					"github.com/wata727/example-3": false,
				},
				IgnoreRule: map[string]bool{
					"aws_instance_invalid_ami":   false,
					"aws_instance_previous_type": true,
				},
				Varfile:          []string{"example3.tfvars"},
				TerraformVersion: "0.12.0",
				Rules: map[string]*Rule{
					"aws_instance_invalid_ami": &Rule{
						Name:    "aws_instance_invalid_ami",
						Enabled: false,
					},
					"aws_instance_previous_type": &Rule{
						Name:    "aws_instance_previous_type",
						Enabled: false,
					},
				},
			},
			Expected: &Config{
				DeepCheck: true, // DeepCheck will not override
				AwsCredentials: client.AwsCredentials{
					AccessKey: "ACCESS_KEY",
					SecretKey: "SECRET_KEY",
					Profile:   "production",
					Region:    "ap-northeast-1",
				},
				IgnoreModule: map[string]bool{
					"github.com/wata727/example-1": true,
					"github.com/wata727/example-2": true,
					"github.com/wata727/example-3": false,
				},
				IgnoreRule: map[string]bool{
					"aws_instance_invalid_type":  false,
					"aws_instance_invalid_ami":   false,
					"aws_instance_previous_type": true,
				},
				Varfile:          []string{"example1.tfvars", "example2.tfvars", "example3.tfvars"},
				TerraformVersion: "0.12.0",
				Rules: map[string]*Rule{
					"aws_instance_invalid_type": &Rule{
						Name:    "aws_instance_invalid_type",
						Enabled: false,
					},
					"aws_instance_invalid_ami": &Rule{
						Name:    "aws_instance_invalid_ami",
						Enabled: false,
					},
					"aws_instance_previous_type": &Rule{
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
		DeepCheck: true,
		AwsCredentials: client.AwsCredentials{
			AccessKey: "access_key",
			SecretKey: "secret_key",
			Region:    "us-east-1",
		},
		IgnoreModule: map[string]bool{
			"github.com/wata727/example-1": true,
			"github.com/wata727/example-2": false,
		},
		IgnoreRule: map[string]bool{
			"aws_instance_invalid_type": false,
			"aws_instance_invalid_ami":  true,
		},
		Varfile:          []string{"example1.tfvars", "example2.tfvars"},
		TerraformVersion: "0.11.1",
		Rules: map[string]*Rule{
			"aws_instance_invalid_type": &Rule{
				Name:    "aws_instance_invalid_type",
				Enabled: false,
			},
			"aws_instance_invalid_ami": &Rule{
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
			Name: "DeepCheck",
			SideEffect: func(c *Config) {
				c.DeepCheck = false
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
			Name: "IgnoreModule",
			SideEffect: func(c *Config) {
				c.IgnoreModule["github.com/wata727/example-1"] = false
			},
		},
		{
			Name: "IgnoreRule",
			SideEffect: func(c *Config) {
				c.IgnoreRule["aws_instance_invalid_type"] = true
			},
		},
		{
			Name: "Varfile",
			SideEffect: func(c *Config) {
				c.Varfile = append(c.Varfile, "new.tfvars")
			},
		},
		{
			Name: "TerraformVersion",
			SideEffect: func(c *Config) {
				c.TerraformVersion = "0.12.0"
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
