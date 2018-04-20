package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
)

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		Name   string
		Input  string
		Result *Config
	}{
		{
			Name:  "load config file",
			Input: "dev_environment",
			Result: &Config{
				DeepCheck: true,
				AwsCredentials: map[string]string{
					"access_key": "AWS_ACCESS_KEY",
					"secret_key": "AWS_SECRET_KEY",
					"region":     "us-east-1",
				},
				IgnoreRule: map[string]bool{
					"aws_instance_invalid_type":  true,
					"aws_instance_previous_type": true,
				},
				IgnoreModule: map[string]bool{
					"github.com/wata727/example-module": true,
				},
				Varfile:            []string{"example1.tfvars", "example2.tfvars"},
				TerraformVersion:   "0.9.11",
				TerraformEnv:       "dev",
				TerraformWorkspace: "dev",
			},
		},
		{
			Name: "empty config file",
			Input: "empty_config",
			Result: Init(),
		},
		{
			Name:   "config file not found",
			Input:  "default",
			Result: Init(),
		},
	}

	prev, _ := filepath.Abs(".")
	dir, _ := os.Getwd()
	defer os.Chdir(prev)

	for _, tc := range cases {
		testDir := dir + "/test-fixtures/" + tc.Input
		os.Chdir(testDir)

		c := Init()
		c.LoadConfig(".tflint.hcl")
		if !reflect.DeepEqual(c, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(c), pp.Sprint(tc.Result), tc.Name)
		}

		os.Chdir(prev)
	}
}

func TestSetAwsCredentials(t *testing.T) {
	type Input struct {
		AccessKey string
		SecretKey string
		Profile   string
		Region    string
	}

	cases := []struct {
		Name   string
		Config *Config
		Input  Input
		Result *Config
	}{
		{
			Name: "set credentials",
			Config: &Config{
				AwsCredentials: map[string]string{},
			},
			Input: Input{
				AccessKey: "AWS_ACCESS_KEY",
				SecretKey: "AWS_SECRET_KEY",
				Profile:   "account1",
				Region:    "us-east-1",
			},
			Result: &Config{
				AwsCredentials: map[string]string{
					"access_key": "AWS_ACCESS_KEY",
					"secret_key": "AWS_SECRET_KEY",
					"profile":    "account1",
					"region":     "us-east-1",
				},
			},
		},
		{
			Name: "do not overwrite",
			Config: &Config{
				AwsCredentials: map[string]string{
					"access_key": "AWS_ACCESS_KEY",
					"secret_key": "AWS_SECRET_KEY",
					"profile":    "account1",
					"region":     "us-east-1",
				},
			},
			Input: Input{
				AccessKey: "",
				SecretKey: "",
				Region:    "",
			},
			Result: &Config{
				AwsCredentials: map[string]string{
					"access_key": "AWS_ACCESS_KEY",
					"secret_key": "AWS_SECRET_KEY",
					"profile":    "account1",
					"region":     "us-east-1",
				},
			},
		},
	}

	for _, tc := range cases {
		tc.Config.SetAwsCredentials(tc.Input.AccessKey, tc.Input.SecretKey, tc.Input.Profile, tc.Input.Region)
		if !reflect.DeepEqual(tc.Config, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(tc.Config), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestHasAwsRegion(t *testing.T) {
	cases := []struct {
		Name   string
		Input  *Config
		Result bool
	}{
		{
			Name: "has region",
			Input: &Config{
				AwsCredentials: map[string]string{
					"region": "us-east-1",
				},
			},
			Result: true,
		},
		{
			Name: "does not have region",
			Input: &Config{
				AwsCredentials: map[string]string{},
			},
			Result: false,
		},
	}

	for _, tc := range cases {
		result := tc.Input.HasAwsRegion()
		if !reflect.DeepEqual(result, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(result), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestHasAwsSharedCredentials(t *testing.T) {
	cases := []struct {
		Name   string
		Input  *Config
		Result bool
	}{
		{
			Name: "has credentials",
			Input: &Config{
				AwsCredentials: map[string]string{
					"profile": "account1",
					"region":  "us-east-1",
				},
			},
			Result: true,
		},
		{
			Name: "does not have profile",
			Input: &Config{
				AwsCredentials: map[string]string{
					"region": "us-east-1",
				},
			},
			Result: false,
		},
		{
			Name: "does not have region",
			Input: &Config{
				AwsCredentials: map[string]string{
					"profile": "account1",
				},
			},
			Result: false,
		},
	}

	for _, tc := range cases {
		result := tc.Input.HasAwsSharedCredentials()
		if !reflect.DeepEqual(result, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(result), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestHasAwsStaticCredentials(t *testing.T) {
	cases := []struct {
		Name   string
		Input  *Config
		Result bool
	}{
		{
			Name: "has credentials",
			Input: &Config{
				AwsCredentials: map[string]string{
					"access_key": "AWS_ACCESS_KEY",
					"secret_key": "AWS_SECRET_KEY",
					"region":     "us-east-1",
				},
			},
			Result: true,
		},
		{
			Name: "does not have access_key",
			Input: &Config{
				AwsCredentials: map[string]string{
					"secret_key": "AWS_SECRET_KEY",
					"region":     "us-east-1",
				},
			},
			Result: false,
		},
		{
			Name: "does not have secret_key",
			Input: &Config{
				AwsCredentials: map[string]string{
					"access_key": "AWS_ACCESS_KEY",
					"region":     "us-east-1",
				},
			},
			Result: false,
		},
		{
			Name: "does not have region",
			Input: &Config{
				AwsCredentials: map[string]string{
					"access_key": "AWS_ACCESS_KEY",
					"secret_key": "AWS_SECRET_KEY",
				},
			},
			Result: false,
		},
	}

	for _, tc := range cases {
		result := tc.Input.HasAwsStaticCredentials()
		if !reflect.DeepEqual(result, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(result), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestSetIgnoreModule(t *testing.T) {
	cases := []struct {
		Name   string
		Config *Config
		Input  string
		Result *Config
	}{
		{
			Name: "set modules",
			Config: &Config{
				IgnoreModule: map[string]bool{},
			},
			Input: "module1,module2",
			Result: &Config{
				IgnoreModule: map[string]bool{
					"module1": true,
					"module2": true,
				},
			},
		},
		{
			Name: "not set",
			Config: &Config{
				IgnoreModule: map[string]bool{},
			},
			Input: "",
			Result: &Config{
				IgnoreModule: map[string]bool{},
			},
		},
		{
			Name: "append modules",
			Config: &Config{
				IgnoreModule: map[string]bool{
					"module1": true,
					"module2": true,
				},
			},
			Input: "module3",
			Result: &Config{
				IgnoreModule: map[string]bool{
					"module1": true,
					"module2": true,
					"module3": true,
				},
			},
		},
	}

	for _, tc := range cases {
		tc.Config.SetIgnoreModule(tc.Input)
		if !reflect.DeepEqual(tc.Config, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(tc.Config), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestSetIgnoreRule(t *testing.T) {
	cases := []struct {
		Name   string
		Config *Config
		Input  string
		Result *Config
	}{
		{
			Name: "set rules",
			Config: &Config{
				IgnoreRule: map[string]bool{},
			},
			Input: "rule1,rule2",
			Result: &Config{
				IgnoreRule: map[string]bool{
					"rule1": true,
					"rule2": true,
				},
			},
		},
		{
			Name: "not set",
			Config: &Config{
				IgnoreRule: map[string]bool{},
			},
			Input: "",
			Result: &Config{
				IgnoreRule: map[string]bool{},
			},
		},
		{
			Name: "append rules",
			Config: &Config{
				IgnoreRule: map[string]bool{
					"rule1": true,
					"rule2": true,
				},
			},
			Input: "rule3",
			Result: &Config{
				IgnoreRule: map[string]bool{
					"rule1": true,
					"rule2": true,
					"rule3": true,
				},
			},
		},
	}

	for _, tc := range cases {
		tc.Config.SetIgnoreRule(tc.Input)
		if !reflect.DeepEqual(tc.Config, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(tc.Config), pp.Sprint(tc.Result), tc.Name)
		}
	}
}

func TestSetVarfile(t *testing.T) {
	cases := []struct {
		Name   string
		Config *Config
		Input  string
		Result *Config
	}{
		{
			Name: "set varfiles",
			Config: &Config{
				Varfile: []string{},
			},
			Input: "example1.tfvars,example2.tfvars",
			Result: &Config{
				Varfile: []string{"terraform.tfvars", "example1.tfvars", "example2.tfvars"},
			},
		},
		{
			Name: "not set",
			Config: &Config{
				Varfile: []string{},
			},
			Input: "",
			Result: &Config{
				Varfile: []string{"terraform.tfvars"},
			},
		},
		{
			Name: "append varfile",
			Config: &Config{
				Varfile: []string{"example1.tfvars"},
			},
			Input: "example2.tfvars",
			Result: &Config{
				Varfile: []string{"terraform.tfvars", "example1.tfvars", "example2.tfvars"},
			},
		},
	}

	for _, tc := range cases {
		tc.Config.SetVarfile(tc.Input)
		if !reflect.DeepEqual(tc.Config, tc.Result) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(tc.Config), pp.Sprint(tc.Result), tc.Name)
		}
	}
}
