package rules

import (
	"errors"
	"reflect"
	"testing"

	"github.com/terraform-linters/tflint/rules/awsrules"
	"github.com/terraform-linters/tflint/rules/terraformrules"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_CheckRuleNames(t *testing.T) {
	cases := []struct {
		Name     string
		Rules    []string
		Expected error
	}{
		{
			Name:     "no error",
			Rules:    []string{"aws_route_not_specified_target"},
			Expected: nil,
		},
		{
			Name: "invalid rule name",
			Rules: []string{
				"aws_route_not_specified_target",
				"invalid_not_exist",
			},
			Expected: errors.New("Rule not found: invalid_not_exist"),
		},
	}

	for _, tc := range cases {
		err := CheckRuleNames(tc.Rules)
		if !reflect.DeepEqual(tc.Expected, err) {
			t.Fatalf("Failed `%s` test: expected `%#v`, but got `%#v`", tc.Name, tc.Expected, err)
		}
	}
}

func Test_NewRules(t *testing.T) {
	// Mock rules in test
	DefaultRules = []Rule{
		awsrules.NewAwsRouteNotSpecifiedTargetRule(),
		terraformrules.NewTerraformNamingConventionRule(),
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
					"terraform_naming_convention": {
						Enabled: true,
					},
				},
			},
			Expected: []Rule{
				awsrules.NewAwsRouteNotSpecifiedTargetRule(),
				terraformrules.NewTerraformNamingConventionRule(),
			},
		},
		{
			Name: "disabled_by_default = true",
			Config: &tflint.Config{
				DisabledByDefault: true,
				Rules: map[string]*tflint.RuleConfig{
					"terraform_naming_convention": {
						Enabled: true,
					},
				},
			},
			Expected: []Rule{
				terraformrules.NewTerraformNamingConventionRule(),
			},
		},
	}

	for _, tc := range cases {
		ret := NewRules(tc.Config)
		if !reflect.DeepEqual(tc.Expected, ret) {
			t.Fatalf("Failed `%s` test: expected rules are `%#v`, but got `%#v`", tc.Name, tc.Expected, ret)
		}
	}
}
