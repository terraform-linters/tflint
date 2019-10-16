package rules

import (
	"errors"
	"reflect"
	"testing"

	"github.com/wata727/tflint/rules/awsrules"
	"github.com/wata727/tflint/rules/terraformrules"
	"github.com/wata727/tflint/tflint"
)

func Test_NewRules(t *testing.T) {
	// Mock rules in test
	DefaultRules = []Rule{
		awsrules.NewAwsRouteNotSpecifiedTargetRule(),
		terraformrules.NewTerraformDashInResourceNameRule(),
	}
	deepCheckRules = []Rule{
		awsrules.NewAwsInstanceInvalidAMIRule(),
	}

	cases := []struct {
		Name     string
		Config   *tflint.Config
		Expected []Rule
		Error    error
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
			Name: "enabled = true",
			Config: &tflint.Config{
				Rules: map[string]*tflint.RuleConfig{
					"terraform_dash_in_resource_name": {
						Enabled: true,
					},
				},
			},
			Expected: []Rule{
				awsrules.NewAwsRouteNotSpecifiedTargetRule(),
				terraformrules.NewTerraformDashInResourceNameRule(),
			},
		},
		{
			Name: "invalid rule name",
			Config: &tflint.Config{
				Rules: map[string]*tflint.RuleConfig{
					"invalid_rule": {
						Enabled: false,
					},
				},
			},
			Expected: []Rule{},
			Error:    errors.New("Rule not found: invalid_rule"),
		},
	}

	for _, tc := range cases {
		ret, err := NewRules(tc.Config)
		if !reflect.DeepEqual(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test: expected rules are `%#v`, but got `%#v`", tc.Name, tc.Expected, ret)
		}
		if err != tc.Error && err.Error() != tc.Error.Error() {
			t.Fatalf("Failed `%s` test: expected error `%#v`, but got `%#v`", tc.Name, tc.Error, err)
		}
	}
}
