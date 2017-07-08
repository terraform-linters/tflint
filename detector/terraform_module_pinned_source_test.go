package detector

import (
	"testing"

	"reflect"

	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
)

func TestDetectTerraformModulePinnedSource(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "git module is not pinned",
			Src: `
module "unpinned" {
  source = "git://hashicorp.com/consul.git"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     "WARNING",
					Message:  "Module source \"git://hashicorp.com/consul.git\" is not pinned",
					Line:     3,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/terraform_module_pinned_source.md",
				},
			},
		},
		{
			Name: "git module reference is default",
			Src: `
module "default git" {
  source = "git://hashicorp.com/consul.git?ref=master"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     "WARNING",
					Message:  "Module source \"git://hashicorp.com/consul.git?ref=master\" uses default ref \"master\"",
					Line:     3,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/terraform_module_pinned_source.md",
				},
			},
		},
		{
			Name: "git module reference is pinned",
			Src: `
module "pinned git" {
  source = "git://hashicorp.com/consul.git?ref=pinned"
}`,
			Issues: []*issue.Issue{},
		},
		{
			Name: "github module is not pinned",
			Src: `
module "unpinned" {
  source = "github.com/hashicorp/consul"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     "WARNING",
					Message:  "Module source \"github.com/hashicorp/consul\" is not pinned",
					Line:     3,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/terraform_module_pinned_source.md",
				},
			},
		},
		{
			Name: "github module reference is default",
			Src: `
module "default git" {
  source = "github.com/hashicorp/consul.git?ref=master"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     "WARNING",
					Message:  "Module source \"github.com/hashicorp/consul.git?ref=master\" uses default ref \"master\"",
					Line:     3,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/terraform_module_pinned_source.md",
				},
			},
		},
		{
			Name: "github module reference is pinned",
			Src: `
module "pinned git" {
  source = "github.com/hashicorp/consul.git?ref=pinned"
}`,
			Issues: []*issue.Issue{},
		},
		{
			Name: "bitbucket module is not pinned",
			Src: `
module "unpinned" {
  source = "bitbucket.org/hashicorp/consul"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     "WARNING",
					Message:  "Module source \"bitbucket.org/hashicorp/consul\" is not pinned",
					Line:     3,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/terraform_module_pinned_source.md",
				},
			},
		},
		{
			Name: "bitbucket module reference is default",
			Src: `
module "default git" {
  source = "bitbucket.org/hashicorp/consul.git?ref=master"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     "WARNING",
					Message:  "Module source \"bitbucket.org/hashicorp/consul.git?ref=master\" uses default ref \"master\"",
					Line:     3,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/terraform_module_pinned_source.md",
				},
			},
		},
		{
			Name: "bitbucket module reference is pinned",
			Src: `
module "pinned git" {
  source = "bitbucket.org/hashicorp/consul.git?ref=pinned"
}`,
			Issues: []*issue.Issue{},
		},
		{
			Name: "mercurial module is not pinned",
			Src: `
module "default mercurial" {
  source = "hg::http://hashicorp.com/consul.hg"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     "WARNING",
					Message:  "Module source \"hg::http://hashicorp.com/consul.hg\" is not pinned",
					Line:     3,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/terraform_module_pinned_source.md",
				},
			},
		},
		{
			Name: "mercurial module reference is default",
			Src: `
module "default mercurial" {
  source = "hg::http://hashicorp.com/consul.hg?rev=default"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "terraform_module_pinned_source",
					Type:     "WARNING",
					Message:  "Module source \"hg::http://hashicorp.com/consul.hg?rev=default\" uses default rev \"default\"",
					Line:     3,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/terraform_module_pinned_source.md",
				},
			},
		},
		{
			Name: "mercurial module reference is pinned",
			Src: `
module "pinned mercurial" {
  source = "hg::http://hashicorp.com/consul.hg?rev=pinned"
}`,
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateTerraformModulePinnedSourceDetector",
			tc.Src,
			"",
			config.Init(),
			config.Init().NewAwsClient(),
			&issues,
		)
		if err != nil {
			t.Fatalf("\nERROR: %s", err)
		}

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(issues), pp.Sprint(tc.Issues), tc.Name)
		}
	}
}
