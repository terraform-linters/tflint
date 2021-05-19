package terraformrules

import (
	"io/ioutil"
	"os"
	"testing"

	hcl "github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint/tflint"
)

func Test_TerraformModulePinnedSource(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Config   string
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
			Name: "git module reference is master",
			Content: `
module "default_git" {
  source = "git://hashicorp.com/consul.git?ref=master"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git://hashicorp.com/consul.git?ref=master\" uses default ref \"master\"",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 55},
					},
				},
			},
		},
		{
			Name: "git module reference is main",
			Content: `
module "default_git" {
  source = "git://hashicorp.com/consul.git?ref=main&depth=99"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git://hashicorp.com/consul.git?ref=main&depth=99\" uses default ref \"main\"",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 62},
					},
				},
			},
		},
		{
			Name: "git module reference is master's-feature-branch",
			Content: `
module "default_git" {
  source = "git://hashicorp.com/consul.git?ref=master's-feature-branch"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "git module reference is pinned",
			Content: `
module "pinned_git" {
  source = "git://hashicorp.com/consul.git?ref=pinned"
}`,
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git://hashicorp.com/consul.git?ref=pinned\" uses a ref which is not a version string",
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
			Expected: tflint.Issues{},
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
			Name: "github module reference is default",
			Content: `
module "default_git" {
  source = "github.com/hashicorp/consul.git?ref=master"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"github.com/hashicorp/consul.git?ref=master\" uses default ref \"master\"",
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
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"github.com/hashicorp/consul.git?ref=pinned\" uses a ref which is not a version string",
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
					Rule:    NewTerraformModulePinnedSourceRule(),
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
			Name: "bitbucket git module reference is default",
			Content: `
module "default_git" {
  source = "bitbucket.org/hashicorp/consul.git?ref=master"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"bitbucket.org/hashicorp/consul.git?ref=master\" uses default ref \"master\"",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 59},
					},
				},
			},
		},
		{
			Name: "bitbucket mercurial module reference is default",
			Content: `
module "default_git" {
  source = "bitbucket.org/hg/mercurial?rev=default"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"bitbucket.org/hg/mercurial?rev=default\" uses default rev \"default\"",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 52},
					},
				},
			},
		},
		{
			Name: "bitbucket git module reference is pinned",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hashicorp/consul.git?ref=pinned"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "bitbucket git module reference is pinned, but style is semver",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hashicorp/consul.git?ref=pinned"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  style = "semver"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"bitbucket.org/hashicorp/consul.git?ref=pinned\" uses a ref which is not a version string",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 59},
					},
				},
			},
		},
		{
			Name: "bitbucket git module reference is pinned to semver",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hashicorp/consul.git?ref=v1.2.3"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  style = "semver"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "bitbucket mercurial module reference is pinned",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hg/mercurial?rev=pinned"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "bitbucket mercurial module reference is pinned, but style is semver",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hg/mercurial?rev=pinned"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  style = "semver"
}`,
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"bitbucket.org/hg/mercurial?rev=pinned\" uses a rev which is not a version string",
					Range: hcl.Range{
						Filename: "module.tf",
						Start:    hcl.Pos{Line: 3, Column: 12},
						End:      hcl.Pos{Line: 3, Column: 51},
					},
				},
			},
		},
		{
			Name: "bitbucket mercurial module reference is pinned to semver",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hg/mercurial?rev=v1.2.3"
}`,
			Config: `
rule "terraform_module_pinned_source" {
  enabled = true
  style = "semver"
}`,
			Expected: tflint.Issues{},
		},
		{
			Name: "bitbucket mercurial module reference is pinned to semver (no leading v)",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hg/mercurial?rev=1.2.3"
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
			Expected: tflint.Issues{
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
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git::https://hashicorp.com/consul.git?ref=master\" uses default ref \"master\"",
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
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"git::ssh://git@github.com/owner/repo.git?ref=master\" uses default ref \"master\"",
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
			Expected: tflint.Issues{},
		},
		{
			Name: "generic git (git::ssh) module reference is pinned",
			Content: `
module "pinned_generic_git_ssh" {
  source = "git::ssh://git@github.com/owner/repo.git?ref=pinned"
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
			Expected: tflint.Issues{
				{
					Rule:    NewTerraformModulePinnedSourceRule(),
					Message: "Module source \"hg::http://hashicorp.com/consul.hg?rev=default\" uses default rev \"default\"",
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
			Expected: tflint.Issues{},
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
			Expected: tflint.Issues{
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
		runner := tflint.TestRunnerWithConfig(t, map[string]string{"module.tf": tc.Content}, loadConfigfromTempFile(t, tc.Config))

		if err := rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		tflint.AssertIssues(t, tc.Expected, runner.Issues)
	}
}

// TODO: Replace with TestRunner
func loadConfigfromTempFile(t *testing.T, content string) *tflint.Config {
	if content == "" {
		return tflint.EmptyConfig()
	}

	tmpfile, err := ioutil.TempFile("", "terraform_module_pinned_source")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	config, err := tflint.LoadConfig(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	return config
}
