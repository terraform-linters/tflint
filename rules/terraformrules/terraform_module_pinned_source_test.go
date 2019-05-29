package terraformrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	version "github.com/hashicorp/go-version"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/project"
	"github.com/wata727/tflint/tflint"
)

func Test_TerraformModulePinnedSource(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected issue.Issues
	}{
		{
			Name: "git module is not pinned",
			Content: `
module "unpinned" {
  source = "git://hashicorp.com/consul.git"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     issue.WARNING,
					Message:  "Module source \"git://hashicorp.com/consul.git\" is not pinned",
					Line:     3,
					File:     "module.tf",
					Link:     project.ReferenceLink("terraform_module_pinned_source"),
				},
			},
		},
		{
			Name: "git module reference is default",
			Content: `
module "default_git" {
  source = "git://hashicorp.com/consul.git?ref=master"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     issue.WARNING,
					Message:  "Module source \"git://hashicorp.com/consul.git?ref=master\" uses default ref \"master\"",
					Line:     3,
					File:     "module.tf",
					Link:     project.ReferenceLink("terraform_module_pinned_source"),
				},
			},
		},
		{
			Name: "git module reference is pinned",
			Content: `
module "pinned_git" {
  source = "git://hashicorp.com/consul.git?ref=pinned"
}`,
			Expected: []*issue.Issue{},
		},
		{
			Name: "github module is not pinned",
			Content: `
module "unpinned" {
  source = "github.com/hashicorp/consul"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     issue.WARNING,
					Message:  "Module source \"github.com/hashicorp/consul\" is not pinned",
					Line:     3,
					File:     "module.tf",
					Link:     project.ReferenceLink("terraform_module_pinned_source"),
				},
			},
		},
		{
			Name: "github module reference is default",
			Content: `
module "default_git" {
  source = "github.com/hashicorp/consul.git?ref=master"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     issue.WARNING,
					Message:  "Module source \"github.com/hashicorp/consul.git?ref=master\" uses default ref \"master\"",
					Line:     3,
					File:     "module.tf",
					Link:     project.ReferenceLink("terraform_module_pinned_source"),
				},
			},
		},
		{
			Name: "github module reference is pinned",
			Content: `
module "pinned_git" {
  source = "github.com/hashicorp/consul.git?ref=pinned"
}`,
			Expected: []*issue.Issue{},
		},
		{
			Name: "bitbucket module is not pinned",
			Content: `
module "unpinned" {
  source = "bitbucket.org/hashicorp/consul"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     issue.WARNING,
					Message:  "Module source \"bitbucket.org/hashicorp/consul\" is not pinned",
					Line:     3,
					File:     "module.tf",
					Link:     project.ReferenceLink("terraform_module_pinned_source"),
				},
			},
		},
		{
			Name: "bitbucket module reference is default",
			Content: `
module "default_git" {
  source = "bitbucket.org/hashicorp/consul.git?ref=master"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     issue.WARNING,
					Message:  "Module source \"bitbucket.org/hashicorp/consul.git?ref=master\" uses default ref \"master\"",
					Line:     3,
					File:     "module.tf",
					Link:     project.ReferenceLink("terraform_module_pinned_source"),
				},
			},
		},
		{
			Name: "bitbucket module reference is pinned",
			Content: `
module "pinned_git" {
  source = "bitbucket.org/hashicorp/consul.git?ref=pinned"
}`,
			Expected: []*issue.Issue{},
		},
		{
			Name: "generic git (git::https) module reference is not pinned",
			Content: `
module "unpinned_generic_git_https" {
  source = "git::https://hashicorp.com/consul.git"
}
`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     issue.WARNING,
					Message:  "Module source \"git::https://hashicorp.com/consul.git\" is not pinned",
					Line:     3,
					File:     "module.tf",
					Link:     project.ReferenceLink("terraform_module_pinned_source"),
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
			Expected: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     issue.WARNING,
					Message:  "Module source \"git::ssh://git@github.com/owner/repo.git\" is not pinned",
					Line:     3,
					File:     "module.tf",
					Link:     project.ReferenceLink("terraform_module_pinned_source"),
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
			Expected: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     issue.WARNING,
					Message:  "Module source \"git::https://hashicorp.com/consul.git?ref=master\" uses default ref \"master\"",
					Line:     3,
					File:     "module.tf",
					Link:     project.ReferenceLink("terraform_module_pinned_source"),
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
			Expected: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     issue.WARNING,
					Message:  "Module source \"git::ssh://git@github.com/owner/repo.git?ref=master\" uses default ref \"master\"",
					Line:     3,
					File:     "module.tf",
					Link:     project.ReferenceLink("terraform_module_pinned_source"),
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
			Expected: []*issue.Issue{},
		},
		{
			Name: "generic git (git::ssh) module reference is pinned",
			Content: `
module "pinned_generic_git_ssh" {
  source = "git::ssh://git@github.com/owner/repo.git?ref=pinned"
}
`,
			Expected: []*issue.Issue{},
		},
		{
			Name: "mercurial module is not pinned",
			Content: `
module "default_mercurial" {
  source = "hg::http://hashicorp.com/consul.hg"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     issue.WARNING,
					Message:  "Module source \"hg::http://hashicorp.com/consul.hg\" is not pinned",
					Line:     3,
					File:     "module.tf",
					Link:     project.ReferenceLink("terraform_module_pinned_source"),
				},
			},
		},
		{
			Name: "mercurial module reference is default",
			Content: `
module "default_mercurial" {
  source = "hg::http://hashicorp.com/consul.hg?rev=default"
}`,
			Expected: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     issue.WARNING,
					Message:  "Module source \"hg::http://hashicorp.com/consul.hg?rev=default\" uses default rev \"default\"",
					Line:     3,
					File:     "module.tf",
					Link:     project.ReferenceLink("terraform_module_pinned_source"),
				},
			},
		},
		{
			Name: "mercurial module reference is pinned",
			Content: `
module "pinned_mercurial" {
  source = "hg::http://hashicorp.com/consul.hg?rev=pinned"
}`,
			Expected: []*issue.Issue{},
		},
	}

	dir, err := ioutil.TempDir("", "TerraformModulePinnedSource")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		loader, err := configload.NewLoader(&configload.Config{})
		if err != nil {
			t.Fatal(err)
		}

		err = ioutil.WriteFile(dir+"/module.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		mod, diags := loader.Parser().LoadConfigDir(dir)
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.ModuleWalkerFunc(func(req *configs.ModuleRequest) (*configs.Module, *version.Version, hcl.Diagnostics) {
			return nil, nil, nil
		}))
		if tfdiags.HasErrors() {
			t.Fatal(tfdiags)
		}

		runner := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
		rule := NewTerraformModulePinnedSourceRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
