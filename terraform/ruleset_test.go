package terraform

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/terraform-linters/tflint-plugin-sdk/hclext"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/rules"
)

func TestApplyConfig(t *testing.T) {
	mustParseExpr := func(input string) hcl.Expression {
		expr, diags := hclsyntax.ParseExpression([]byte(input), "", hcl.InitialPos)
		if diags.HasErrors() {
			panic(diags)
		}
		return expr
	}

	tests := []struct {
		name   string
		global *tflint.Config
		config *hclext.BodyContent
		want   []string
	}{
		{
			name:   "default",
			global: &tflint.Config{},
			config: &hclext.BodyContent{},
			want: []string{
				"terraform_comment_syntax",
				"terraform_deprecated_index",
				"terraform_deprecated_interpolation",
			},
		},
		{
			name:   "disabled by default",
			global: &tflint.Config{DisabledByDefault: true},
			config: &hclext.BodyContent{},
			want:   []string{},
		},
		{
			name:   "preset",
			global: &tflint.Config{},
			config: &hclext.BodyContent{
				Attributes: hclext.Attributes{
					"preset": &hclext.Attribute{Name: "preset", Expr: mustParseExpr(`"recommended"`)},
				},
			},
			want: []string{
				"terraform_comment_syntax",
				"terraform_deprecated_index",
			},
		},
		{
			name: "rule config",
			global: &tflint.Config{
				Rules: map[string]*tflint.RuleConfig{
					"terraform_comment_syntax": {
						Name:    "terraform_comment_syntax",
						Enabled: false,
					},
				},
			},
			config: &hclext.BodyContent{},
			want: []string{
				"terraform_deprecated_index",
				"terraform_deprecated_interpolation",
			},
		},
		{
			name:   "disabled by default + preset",
			global: &tflint.Config{DisabledByDefault: true},
			config: &hclext.BodyContent{
				Attributes: hclext.Attributes{
					"preset": &hclext.Attribute{Name: "preset", Expr: mustParseExpr(`"recommended"`)},
				},
			},
			want: []string{
				"terraform_comment_syntax",
				"terraform_deprecated_index",
			},
		},
		{
			name: "disabled by default + rule config",
			global: &tflint.Config{
				Rules: map[string]*tflint.RuleConfig{
					"terraform_comment_syntax": {
						Name:    "terraform_comment_syntax",
						Enabled: true,
					},
				},
				DisabledByDefault: true,
			},
			config: &hclext.BodyContent{},
			want: []string{
				"terraform_comment_syntax",
			},
		},
		{
			name: "preset + rule config",
			global: &tflint.Config{
				Rules: map[string]*tflint.RuleConfig{
					"terraform_comment_syntax": {
						Name:    "terraform_comment_syntax",
						Enabled: false,
					},
				},
			},
			config: &hclext.BodyContent{
				Attributes: hclext.Attributes{
					"preset": &hclext.Attribute{Name: "preset", Expr: mustParseExpr(`"recommended"`)},
				},
			},
			want: []string{
				"terraform_deprecated_index",
			},
		},
		{
			name: "disabled by default + preset + rule config",
			global: &tflint.Config{
				Rules: map[string]*tflint.RuleConfig{
					"terraform_comment_syntax": {
						Name:    "terraform_comment_syntax",
						Enabled: false,
					},
				},
				DisabledByDefault: true,
			},
			config: &hclext.BodyContent{
				Attributes: hclext.Attributes{
					"preset": &hclext.Attribute{Name: "preset", Expr: mustParseExpr(`"recommended"`)},
				},
			},
			want: []string{
				"terraform_deprecated_index",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ruleset := &RuleSet{
				PresetRules: map[string][]tflint.Rule{
					"all": {
						rules.NewTerraformCommentSyntaxRule(),
						rules.NewTerraformDeprecatedIndexRule(),
						rules.NewTerraformDeprecatedInterpolationRule(),
					},
					"recommended": {
						rules.NewTerraformCommentSyntaxRule(),
						rules.NewTerraformDeprecatedIndexRule(),
					},
				},
				rulesetConfig: &Config{},
			}

			err := ruleset.ApplyGlobalConfig(test.global)
			if err != nil {
				t.Fatal(err)
			}

			err = ruleset.ApplyConfig(test.config)
			if err != nil {
				t.Fatal(err)
			}

			got := make([]string, len(ruleset.EnabledRules))
			for i, r := range ruleset.EnabledRules {
				got[i] = r.Name()
			}

			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
