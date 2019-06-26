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
		awsrules.NewAwsRouteNotSpecifiedTargetRule(),
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
				awsrules.NewAwsRouteNotSpecifiedTargetRule(),
			},
		},
		{
			Name: "enabled deep check",
			Config: &tflint.Config{
				DeepCheck: true,
			},
			Expected: []Rule{
				awsrules.NewAwsRouteNotSpecifiedTargetRule(),
				awsrules.NewAwsInstanceInvalidAMIRule(),
			},
		},
		{
			Name: "ignore_rule",
			Config: &tflint.Config{
				IgnoreRule: map[string]bool{
					"aws_route_not_specified_target": true,
				},
			},
			Expected: []Rule{},
		},
		{
			Name: "enabled = false",
			Config: &tflint.Config{
				Rules: map[string]*tflint.RuleConfig{
					"aws_route_not_specified_target": {
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
					"aws_route_not_specified_target": true,
				},
				Rules: map[string]*tflint.RuleConfig{
					"aws_route_not_specified_target": {
						Enabled: true,
					},
				},
			},
			Expected: []Rule{
				awsrules.NewAwsRouteNotSpecifiedTargetRule(),
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
