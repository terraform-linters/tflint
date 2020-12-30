package rules

import (
	"errors"
	"reflect"
	"testing"

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
			Rules:    []string{"terraform_deprecated_interpolation"},
			Expected: nil,
		},
		{
			Name: "invalid rule name",
			Rules: []string{
				"terraform_deprecated_interpolation",
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
		terraformrules.NewTerraformDeprecatedInterpolationRule(),
		terraformrules.NewTerraformNamingConventionRule(),
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
				terraformrules.NewTerraformDeprecatedInterpolationRule(),
			},
		},
		{
			Name: "enabled = false",
			Config: &tflint.Config{
				Rules: map[string]*tflint.RuleConfig{
					"terraform_deprecated_interpolation": {
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
				terraformrules.NewTerraformDeprecatedInterpolationRule(),
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
