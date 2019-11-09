package terraformrules

import (
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformModulePinnedSourceIsSemver(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected tflint.Issues
	}{
		{
			Name: "git module is not pinned",
			Content: `
module "unpinned" {
  source = "git://hashicorp.com/consul.git"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModuleSemverSourceRule(),
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
			Name: "git module reference is not semver",
			Content: `
module "default_git" {
  source = "git://hashicorp.com/consul.git?ref=master"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModuleSemverSourceRule(),
					Message: "Module source \"git://hashicorp.com/consul.git?ref=master\" uses a ref which is not a version string",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 55},
					},
				},
			},
		},
		{
			Name: "git module reference is semver",
			Content: `
module "pinned_git" {
  source = "git://hashicorp.com/consul.git?ref=v1.2.3"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "git module reference is semver (no leading v)",
			Content: `
module "pinned_git" {
  source = "git://hashicorp.com/consul.git?ref=1.2.3"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "github module is not pinned",
			Content: `
module "unpinned" {
  source = "github.com/hashicorp/consul"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModuleSemverSourceRule(),
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
			Name: "github module reference is not semver",
			Content: `
module "default_git" {
  source = "github.com/hashicorp/consul.git?ref=master"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModuleSemverSourceRule(),
					Message: "Module source \"github.com/hashicorp/consul.git?ref=master\" uses a ref which is not a version string",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 56},
					},
				},
			},
		},
		{
			Name: "github module reference is semver",
			Content: `
module "pinned_git" {
  source = "github.com/hashicorp/consul.git?ref=v1.2.3"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "github module reference is semver (no leading v)",
			Content: `
module "pinned_git" {
  source = "github.com/hashicorp/consul.git?ref=v1.2.3"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "bitbucket module is not pinned",
			Content: `
module "unpinned" {
  source = "bitbucket.org/hashicorp/consul"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModuleSemverSourceRule(),
					Message: "Module source \"bitbucket.org/hashicorp/consul\" is not pinned",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 44},
					},
				},
			},
		},
		{
			Name: "bitbucket module reference is not semver",
			Content: `
module "default_git" {
  source = "bitbucket.org/hashicorp/consul.git?ref=master"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModuleSemverSourceRule(),
					Message: "Module source \"bitbucket.org/hashicorp/consul.git?ref=master\" uses a ref which is not a version string",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 59},
					},
				},
			},
		},
		{
			Name: "bitbucket Git module reference is pinned to semver",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hashicorp/consul.git?ref=v1.2.3"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "bitbucket Git module reference is pinned to semver (no leading v)",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hashicorp/consul.git?ref=v1.2.3"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "generic git (git::https) module reference is not pinned",
			Content: `
module "unpinned_generic_git_https" {
  source = "git::https://hashicorp.com/consul.git"
}
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModuleSemverSourceRule(),
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
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModuleSemverSourceRule(),
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
			Name: "generic git (git::https) module reference is not semver",
			Content: `
module "default_generic_git_https" {
  source = "git::https://hashicorp.com/consul.git?ref=master"
}
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModuleSemverSourceRule(),
					Message: "Module source \"git::https://hashicorp.com/consul.git?ref=master\" uses a ref which is not a version string",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 62},
					},
				},
			},
		},
		{
			Name: "generic git (git::ssh) module reference is not semver",
			Content: `
module "default_generic_git_ssh" {
  source = "git::ssh://git@github.com/owner/repo.git?ref=master"
}
`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModuleSemverSourceRule(),
					Message: "Module source \"git::ssh://git@github.com/owner/repo.git?ref=master\" uses a ref which is not a version string",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 65},
					},
				},
			},
		},
		{
			Name: "generic git (git::https) module reference is semver",
			Content: `
module "pinned_generic_git_https" {
  source = "git::https://hashicorp.com/consul.git?ref=v1.2.3"
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "generic git (git::ssh) module reference is semver",
			Content: `
module "pinned_generic_git_ssh" {
  source = "git::ssh://git@github.com/owner/repo.git?ref=v12.1.16"
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "generic git (git::https) module reference is semver (no leading v)",
			Content: `
module "pinned_generic_git_https" {
  source = "git::https://hashicorp.com/consul.git?ref=1.2.3"
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "generic git (git::ssh) module reference is semver (no leading v)",
			Content: `
module "pinned_generic_git_ssh" {
  source = "git::ssh://git@github.com/owner/repo.git?ref=12.1.16"
}
`,
			Expected: tflint.Issues{},
		},
		{
			Name: "mercurial module is not pinned",
			Content: `
module "default_mercurial" {
  source = "hg::http://hashicorp.com/consul.hg"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModuleSemverSourceRule(),
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
			Name: "mercurial module reference is not semver",
			Content: `
module "default_mercurial" {
  source = "hg::http://hashicorp.com/consul.hg?rev=default"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModuleSemverSourceRule(),
					Message: "Module source \"hg::http://hashicorp.com/consul.hg?rev=default\" uses a rev which is not a version string",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 60},
					},
				},
			},
		},
		{
			Name: "mercurial module reference is semver",
			Content: `
module "pinned_mercurial" {
  source = "hg::http://hashicorp.com/consul.hg?rev=v1.2.3"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "mercurial module reference is semver (no leading v)",
			Content: `
module "pinned_mercurial" {
  source = "hg::http://hashicorp.com/consul.hg?rev=1.2.3"
}`,
			Expected: tflint.Issues{},
		},
	}

	rule := NewTerraformModuleSemverSourceRule()

	for _, tc := range cases {
		runner := tflint.TestRunner(t, map[string]string{"module.tf": tc.Content})

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}
