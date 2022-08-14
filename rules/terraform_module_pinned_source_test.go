package rules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TerraformModulePinnedSource(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Config   string
		Expected helper.Issues
	}{
		{
			Name: "local module",
			Content: `
module "unpinned" {
  source = "./local"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "git module is not pinned",
			Content: `
module "unpinned" {
  source = "git://hashicorp.com/consul.git"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git://hashicorp.com/consul.git\" is not pinned",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 44},
					},
				},
			},
		},
		{
			Name: "git module reference is default",
			Content: `
module "default_git" {
  source = "git://hashicorp.com/consul.git?ref=master"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git://hashicorp.com/consul.git?ref=master\" uses a default branch as ref (master)",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 55},
					},
				},
			},
		},
		{
			Name: "git module reference is pinned",
			Content: `
module "pinned_git" {
  source = "git://hashicorp.com/consul.git?ref=pinned"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "invalid URL",
			Content: `
module "invalid" {
  source = "git://#{}.com"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: `Module source "git://#{}.com" is not a valid URL`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 27},
					},
				},
			},
		},
		{
			Name: "git module reference is pinned, but style is semver",
			Content: `
module "pinned_git" {
  source = "git://hashicorp.com/consul.git?ref=pinned"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  style = "semver"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git://hashicorp.com/consul.git?ref=pinned\" uses a ref which is not a semantic version string",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 55},
					},
				},
			},
		},
		{
			Name: "git module reference is pinned to semver",
			Content: `
module "pinned_git" {
  source = "git://hashicorp.com/consul.git?ref=v1.2.3"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  style = "semver"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "git module reference is pinned to semver (no leading v)",
			Content: `
module "pinned_git" {
  source = "git://hashicorp.com/consul.git?ref=1.2.3"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  style = "semver"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "github module is not pinned",
			Content: `
module "unpinned" {
  source = "github.com/hashicorp/consul"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"github.com/hashicorp/consul\" is not pinned",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 41},
					},
				},
			},
		},
		{
			Name: "github ssh module is not pinned",
			Content: `
module "unpinned" {
  source = "git@github.com:hashicorp/consul.git"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git@github.com:hashicorp/consul.git\" is not pinned",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 49},
					},
				},
			},
		},
		{
			Name: "github module reference is default",
			Content: `
module "default_git" {
  source = "github.com/hashicorp/consul.git?ref=master"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"github.com/hashicorp/consul.git?ref=master\" uses a default branch as ref (master)",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 56},
					},
				},
			},
		},
		{
			Name: "github module reference is pinned",
			Content: `
module "pinned_git" {
  source = "github.com/hashicorp/consul.git?ref=pinned"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "github ssh module is pinned",
			Content: `
module "unpinned" {
  source = "git@github.com:hashicorp/consul.git?ref=pinned"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "github module reference is pinned, but style is semver",
			Content: `
module "pinned_git" {
  source = "github.com/hashicorp/consul.git?ref=pinned"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  style = "semver"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"github.com/hashicorp/consul.git?ref=pinned\" uses a ref which is not a semantic version string",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 56},
					},
				},
			},
		},
		{
			Name: "github module reference is pinned to semver",
			Content: `
module "pinned_git" {
  source = "github.com/hashicorp/consul.git?ref=v1.2.3"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  style = "semver"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "bitbucket module is not pinned",
			Content: `
module "unpinned" {
  source = "bitbucket.org/hashicorp/tf-test-git"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"bitbucket.org/hashicorp/tf-test-git\" is not pinned",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 49},
					},
				},
			},
		},
		{
			Name: "bitbucket git module reference is default",
			Content: `
module "default_git" {
  source = "bitbucket.org/hashicorp/tf-test-git.git?ref=master"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"bitbucket.org/hashicorp/tf-test-git.git?ref=master\" uses a default branch as ref (master)",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 64},
					},
				},
			},
		},
		{
			Name: "bitbucket git module reference is pinned",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hashicorp/tf-test-git.git?ref=pinned"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "bitbucket git module reference is pinned, but style is semver",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hashicorp/tf-test-git.git?ref=pinned"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  style = "semver"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"bitbucket.org/hashicorp/tf-test-git.git?ref=pinned\" uses a ref which is not a semantic version string",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 64},
					},
				},
			},
		},
		{
			Name: "bitbucket git module reference is pinned to semver",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hashicorp/tf-test-git.git?ref=v1.2.3"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  style = "semver"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "generic git (git::https) module reference is not pinned",
			Content: `
module "unpinned_generic_git_https" {
  source = "git::https://hashicorp.com/consul.git"
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git::https://hashicorp.com/consul.git\" is not pinned",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 51},
					},
				},
			},
		},
		{
			Name: "generic git (git::ssh) module reference is not pinned",
			Content: `
module "unpinned_generic_git_ssh" {
  source = "git::ssh://git@github.com/owner/repo.git"
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git::ssh://git@github.com/owner/repo.git\" is not pinned",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 54},
					},
				},
			},
		},
		{
			Name: "generic git (git::https) module reference is default",
			Content: `
module "default_generic_git_https" {
  source = "git::https://hashicorp.com/consul.git?ref=master"
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git::https://hashicorp.com/consul.git?ref=master\" uses a default branch as ref (master)",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 62},
					},
				},
			},
		},
		{
			Name: "generic git (git::ssh) module reference is default",
			Content: `
module "default_generic_git_ssh" {
  source = "git::ssh://git@github.com/owner/repo.git?ref=master"
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git::ssh://git@github.com/owner/repo.git?ref=master\" uses a default branch as ref (master)",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 65},
					},
				},
			},
		},
		{
			Name: "generic git (git::https) module reference is pinned",
			Content: `
module "pinned_generic_git_https" {
  source = "git::https://hashicorp.com/consul.git?ref=pinned"
}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "generic git (git::ssh) module reference is pinned",
			Content: `
module "pinned_generic_git_ssh" {
  source = "git::ssh://git@github.com/owner/repo.git?ref=pinned"
}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "github module reference is unpinned via custom branches",
			Content: `
module "pinned_git" {
  source = "github.com/hashicorp/consul.git?ref=foo"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  default_branches = ["foo"]
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"github.com/hashicorp/consul.git?ref=foo\" uses a default branch as ref (foo)",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 53},
					},
				},
			},
		},
		{
			Name: "mercurial module is not pinned",
			Content: `
module "default_mercurial" {
  source = "hg::http://hashicorp.com/consul.hg"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"hg::http://hashicorp.com/consul.hg\" is not pinned",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 48},
					},
				},
			},
		},
		{
			Name: "mercurial module reference is default",
			Content: `
module "default_mercurial" {
  source = "hg::http://hashicorp.com/consul.hg?rev=default"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"hg::http://hashicorp.com/consul.hg?rev=default\" uses a default branch as rev (default)",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 60},
					},
				},
			},
		},
		{
			Name: "mercurial module reference is pinned",
			Content: `
module "pinned_mercurial" {
  source = "hg::http://hashicorp.com/consul.hg?rev=pinned"
}`,
			Expected: helper.Issues{},
		},
		{
			Name: "git module is not pinned with default config",
			Content: `
module "unpinned" {
  source = "git://hashicorp.com/consul.git"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  style = "flexible"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git://hashicorp.com/consul.git\" is not pinned",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 44},
					},
				},
			},
		},
	}

	rule := NewTerraformModulePinnedSourceRule()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := helper.TestRunner(t, map[string]string{"module.tf": tc.Content, ".tflint.hcl": tc.Config})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Expected, runner.Issues)
		})
	}
}
