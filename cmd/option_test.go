package cmd

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	flags "github.com/jessevdk/go-flags"
	"github.com/wata727/tflint/client"
	"github.com/wata727/tflint/tflint"
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
			Name:    "--deep",
			Command: "./tflint --deep",
			Expected: &tflint.Config{
				DeepCheck:        true,
				AwsCredentials:   client.AwsCredentials{},
				IgnoreModule:     map[string]bool{},
				IgnoreRule:       map[string]bool{},
				Varfile:          []string{},
				TerraformVersion: "",
				Rules:            map[string]*tflint.Rule{},
			},
		},
		{
			Name:    "AWS static credentials",
			Command: "./tflint --aws-access-key AWS_ACCESS_KEY_ID --aws-secret-key AWS_SECRET_ACCESS_KEY --aws-region us-east-1",
			Expected: &tflint.Config{
				DeepCheck: false,
				AwsCredentials: client.AwsCredentials{
					AccessKey: "AWS_ACCESS_KEY_ID",
					SecretKey: "AWS_SECRET_ACCESS_KEY",
					Region:    "us-east-1",
				},
				IgnoreModule:     map[string]bool{},
				IgnoreRule:       map[string]bool{},
				Varfile:          []string{},
				TerraformVersion: "",
				Rules:            map[string]*tflint.Rule{},
			},
		},
		{
			Name:    "AWS shared credentials",
			Command: "./tflint --aws-profile production --aws-region us-east-1",
			Expected: &tflint.Config{
				DeepCheck: false,
				AwsCredentials: client.AwsCredentials{
					Profile: "production",
					Region:  "us-east-1",
				},
				IgnoreModule:     map[string]bool{},
				IgnoreRule:       map[string]bool{},
				Varfile:          []string{},
				TerraformVersion: "",
				Rules:            map[string]*tflint.Rule{},
			},
		},
		{
			Name:    "--ignore-module",
			Command: "./tflint --ignore-module module1,module2",
			Expected: &tflint.Config{
				DeepCheck:        false,
				AwsCredentials:   client.AwsCredentials{},
				IgnoreModule:     map[string]bool{"module1": true, "module2": true},
				IgnoreRule:       map[string]bool{},
				Varfile:          []string{},
				TerraformVersion: "",
				Rules:            map[string]*tflint.Rule{},
			},
		},
		{
			Name:    "--ignore-rule",
			Command: "./tflint --ignore-rule rule1,rule2",
			Expected: &tflint.Config{
				DeepCheck:        false,
				AwsCredentials:   client.AwsCredentials{},
				IgnoreModule:     map[string]bool{},
				IgnoreRule:       map[string]bool{"rule1": true, "rule2": true},
				Varfile:          []string{},
				TerraformVersion: "",
				Rules:            map[string]*tflint.Rule{},
			},
		},
		{
			Name:    "--var-file",
			Command: "./tflint --var-file example1.tfvars,example2.tfvars",
			Expected: &tflint.Config{
				DeepCheck:        false,
				AwsCredentials:   client.AwsCredentials{},
				IgnoreModule:     map[string]bool{},
				IgnoreRule:       map[string]bool{},
				Varfile:          []string{"example1.tfvars", "example2.tfvars"},
				TerraformVersion: "",
				Rules:            map[string]*tflint.Rule{},
			},
		},
		{
			Name:    "--fast",
			Command: "./tflint --fast",
			Expected: &tflint.Config{
				DeepCheck:        false,
				AwsCredentials:   client.AwsCredentials{},
				IgnoreModule:     map[string]bool{},
				IgnoreRule:       map[string]bool{"aws_instance_invalid_ami": true},
				Varfile:          []string{},
				TerraformVersion: "",
				Rules:            map[string]*tflint.Rule{},
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
