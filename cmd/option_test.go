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
			Name:    "--module",
			Command: "./tflint --module",
			Expected: &tflint.Config{
				Module:         true,
				DeepCheck:      false,
				Force:          false,
				AwsCredentials: client.AwsCredentials{},
				IgnoreModule:   map[string]bool{},
				IgnoreRule:     map[string]bool{},
				Varfile:        []string{},
				Variables:      []string{},
				Rules:          map[string]*tflint.RuleConfig{},
			},
		},
		{
			Name:    "--deep",
			Command: "./tflint --deep",
			Expected: &tflint.Config{
				Module:         false,
				DeepCheck:      true,
				Force:          false,
				AwsCredentials: client.AwsCredentials{},
				IgnoreModule:   map[string]bool{},
				IgnoreRule:     map[string]bool{},
				Varfile:        []string{},
				Variables:      []string{},
				Rules:          map[string]*tflint.RuleConfig{},
			},
		},
		{
			Name:    "--force",
			Command: "./tflint --force",
			Expected: &tflint.Config{
				Module:         false,
				DeepCheck:      false,
				Force:          true,
				AwsCredentials: client.AwsCredentials{},
				IgnoreModule:   map[string]bool{},
				IgnoreRule:     map[string]bool{},
				Varfile:        []string{},
				Variables:      []string{},
				Rules:          map[string]*tflint.RuleConfig{},
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
				IgnoreModule: map[string]bool{},
				IgnoreRule:   map[string]bool{},
				Varfile:      []string{},
				Variables:    []string{},
				Rules:        map[string]*tflint.RuleConfig{},
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
				IgnoreModule: map[string]bool{},
				IgnoreRule:   map[string]bool{},
				Varfile:      []string{},
				Variables:    []string{},
				Rules:        map[string]*tflint.RuleConfig{},
			},
		},
		{
			Name:    "--ignore-module",
			Command: "./tflint --ignore-module module1,module2",
			Expected: &tflint.Config{
				Module:         false,
				DeepCheck:      false,
				Force:          false,
				AwsCredentials: client.AwsCredentials{},
				IgnoreModule:   map[string]bool{"module1": true, "module2": true},
				IgnoreRule:     map[string]bool{},
				Varfile:        []string{},
				Variables:      []string{},
				Rules:          map[string]*tflint.RuleConfig{},
			},
		},
		{
			Name:    "--ignore-rule",
			Command: "./tflint --ignore-rule rule1,rule2",
			Expected: &tflint.Config{
				Module:         false,
				DeepCheck:      false,
				Force:          false,
				AwsCredentials: client.AwsCredentials{},
				IgnoreModule:   map[string]bool{},
				IgnoreRule:     map[string]bool{"rule1": true, "rule2": true},
				Varfile:        []string{},
				Variables:      []string{},
				Rules:          map[string]*tflint.RuleConfig{},
			},
		},
		{
			Name:    "--var-file",
			Command: "./tflint --var-file example1.tfvars,example2.tfvars",
			Expected: &tflint.Config{
				Module:         false,
				DeepCheck:      false,
				Force:          false,
				AwsCredentials: client.AwsCredentials{},
				IgnoreModule:   map[string]bool{},
				IgnoreRule:     map[string]bool{},
				Varfile:        []string{"example1.tfvars", "example2.tfvars"},
				Variables:      []string{},
				Rules:          map[string]*tflint.RuleConfig{},
			},
		},
		{
			Name:    "--var",
			Command: "./tflint --var foo=bar --var bar=baz",
			Expected: &tflint.Config{
				Module:         false,
				DeepCheck:      false,
				Force:          false,
				AwsCredentials: client.AwsCredentials{},
				IgnoreModule:   map[string]bool{},
				IgnoreRule:     map[string]bool{},
				Varfile:        []string{},
				Variables:      []string{"foo=bar", "bar=baz"},
				Rules:          map[string]*tflint.RuleConfig{},
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
