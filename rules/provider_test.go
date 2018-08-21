package rules

import (
	"reflect"
	"testing"

	"github.com/wata727/tflint/rules/awsrules"
	"github.com/wata727/tflint/tflint"
)

func Test_NewRules(t *testing.T) {
	// Mock rules in test
	DefaultRules = []Rule{
		awsrules.NewAwsInstanceInvalidTypeRule(),
	}
	deepCheckRules = []Rule{
		awsrules.NewAwsInstanceInvalidAMIRule(),
	}

	cases := []struct {
		Name     string
		Config   *tflint.Config
		Expected []Rule
	}{
		{
			Name:   "default",
			Config: tflint.EmptyConfig(),
			Expected: []Rule{
				awsrules.NewAwsInstanceInvalidTypeRule(),
			},
		},
		{
			Name: "enabled deep check",
			Config: &tflint.Config{
				DeepCheck: true,
			},
			Expected: []Rule{
				awsrules.NewAwsInstanceInvalidTypeRule(),
				awsrules.NewAwsInstanceInvalidAMIRule(),
			},
		},
		{
			Name: "ignore_rule",
			Config: &tflint.Config{
				IgnoreRule: map[string]bool{
					"aws_instance_invalid_type": true,
				},
			},
			Expected: []Rule{},
		},
		{
			Name: "enabled = false",
			Config: &tflint.Config{
				Rules: map[string]*tflint.Rule{
					"aws_instance_invalid_type": {
						Enabled: false,
					},
				},
			},
			Expected: []Rule{},
		},
		{
			Name: "`enabled = true` and `ignore_rule`",
			Config: &tflint.Config{
				IgnoreRule: map[string]bool{
					"aws_instance_invalid_type": true,
				},
				Rules: map[string]*tflint.Rule{
					"aws_instance_invalid_type": {
						Enabled: true,
					},
				},
			},
			Expected: []Rule{
				awsrules.NewAwsInstanceInvalidTypeRule(),
			},
		},
	}

	for _, tc := range cases {
		ret := NewRules(tc.Config)
		if !reflect.DeepEqual(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test: expected rules are `%#v`, but get `%#v`", tc.Name, tc.Expected, ret)
		}
	}
}
