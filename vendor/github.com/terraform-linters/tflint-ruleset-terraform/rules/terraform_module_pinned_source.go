package rules

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/hashicorp/go-getter"
	"github.com/terraform-linters/tflint-plugin-sdk/tflint"
	"github.com/terraform-linters/tflint-ruleset-terraform/project"
	"github.com/terraform-linters/tflint-ruleset-terraform/terraform"
)

// TerraformModulePinnedSourceRule checks unpinned or default version module source
type TerraformModulePinnedSourceRule struct {
	tflint.DefaultRule

	attributeName string
}

type terraformModulePinnedSourceRuleConfig struct {
	Style           string   `hclext:"style,optional"`
	DefaultBranches []string `hclext:"default_branches,optional"`
}

// NewTerraformModulePinnedSourceRule returns new rule with default attributes
func NewTerraformModulePinnedSourceRule() *TerraformModulePinnedSourceRule {
	return &TerraformModulePinnedSourceRule{
		attributeName: "source",
	}
}

// Name returns the rule name
func (r *TerraformModulePinnedSourceRule) Name() string {
	return "terraform_module_pinned_source"
}

// Enabled returns whether the rule is enabled by default
func (r *TerraformModulePinnedSourceRule) Enabled() bool {
	return true
}

// Severity returns the rule severity
func (r *TerraformModulePinnedSourceRule) Severity() tflint.Severity {
	return tflint.WARNING
}

// Link returns the rule reference link
func (r *TerraformModulePinnedSourceRule) Link() string {
	return project.ReferenceLink(r.Name())
}

// Check checks if module source version is pinned
// Note that this rule is valid only for Git or Mercurial source
func (r *TerraformModulePinnedSourceRule) Check(rr tflint.Runner) error {
	runner := rr.(*terraform.Runner)

	path, err := runner.GetModulePath()
	if err != nil {
		return err
	}
	if !path.IsRoot() {
		// This rule does not evaluate child modules.
		return nil
	}

	config := terraformModulePinnedSourceRuleConfig{Style: "flexible"}
	config.DefaultBranches = append(config.DefaultBranches, "master", "main", "default", "develop")
	if err := runner.DecodeRuleConfig(r.Name(), &config); err != nil {
		return err
	}

	calls, diags := runner.GetModuleCalls()
	if diags.HasErrors() {
		return diags
	}

	for _, call := range calls {
		if err := r.checkModule(runner, call, config); err != nil {
			return err
		}
	}

	return nil
}

func (r *TerraformModulePinnedSourceRule) checkModule(runner tflint.Runner, module *terraform.ModuleCall, config terraformModulePinnedSourceRuleConfig) error {
	source, err := getter.Detect(module.Source, filepath.Dir(module.DefRange.Filename), []getter.Detector{
		// https://github.com/hashicorp/terraform/blob/51b0aee36cc2145f45f5b04051a01eb6eb7be8bf/internal/getmodules/getter.go#L30-L52
		new(getter.GitHubDetector),
		new(getter.GitDetector),
		new(getter.BitBucketDetector),
		new(getter.GCSDetector),
		new(getter.S3Detector),
		new(getter.FileDetector),
	})
	if err != nil {
		return err
	}

	u, err := url.Parse(source)
	if err != nil {
		return err
	}

	switch u.Scheme {
	case "git", "hg":
	default:
		return nil
	}

	if u.Opaque != "" {
		// for git:: or hg:: pseudo-URLs, Opaque is :https, but query will still be parsed
		query := u.RawQuery
		u, err = url.Parse(strings.TrimPrefix(u.Opaque, ":"))
		if err != nil {
			return err
		}

		u.RawQuery = query
	}

	if u.Hostname() == "" {
		return runner.EmitIssue(
			r,
			fmt.Sprintf("Module source %q is not a valid URL", module.Source),
			module.SourceAttr.Expr.Range(),
		)
	}

	query := u.Query()

	if ref := query.Get("ref"); ref != "" {
		return r.checkRevision(runner, module, config, "ref", ref)
	}

	if rev := query.Get("rev"); rev != "" {
		return r.checkRevision(runner, module, config, "rev", rev)
	}

	return runner.EmitIssue(
		r,
		fmt.Sprintf(`Module source "%s" is not pinned`, module.Source),
		module.SourceAttr.Expr.Range(),
	)
}

func (r *TerraformModulePinnedSourceRule) checkRevision(runner tflint.Runner, module *terraform.ModuleCall, config terraformModulePinnedSourceRuleConfig, key string, value string) error {
	switch config.Style {
	// The "flexible" style requires a revision that is not a default branch
	case "flexible":
		for _, branch := range config.DefaultBranches {
			if value == branch {
				return runner.EmitIssue(
					r,
					fmt.Sprintf("Module source \"%s\" uses a default branch as %s (%s)", module.Source, key, branch),
					module.SourceAttr.Expr.Range(),
				)
			}
		}
	// The "semver" style requires a revision that is a semantic version
	case "semver":
		_, err := semver.NewVersion(value)
		if err != nil {
			return runner.EmitIssue(
				r,
				fmt.Sprintf("Module source \"%s\" uses a %s which is not a semantic version string", module.Source, key),
				module.SourceAttr.Expr.Range(),
			)
		}
	default:
		return fmt.Errorf("`%s` is invalid style", config.Style)
	}

	return nil
}
