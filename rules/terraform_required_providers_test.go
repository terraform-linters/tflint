package rules

import (
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/terraform-linters/tflint-plugin-sdk/helper"
)

func Test_TerraformRequiredProvidersRule(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected helper.Issues
	}{
		{
			Name: "no version",
			Content: `
provider "template" {}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `Missing version constraint for provider "template" in "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 20,
						},
					},
				},
			},
		},
		{
			Name: "implicit provider - resource",
			Content: `
resource "random_string" "foo" {
	length = 16
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `Missing version constraint for provider "random" in "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 31,
						},
					},
				},
			},
		},
		{
			Name: "implicit provider - data source",
			Content: `
data "template_file" "foo" {
	template = ""
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `Missing version constraint for provider "template" in "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 27,
						},
					},
				},
			},
		},
		{
			Name: "required_providers object",
			Content: `
terraform {
	required_providers {
		template = {
			source  = "hashicorp/template"
			version = "~> 2" 
		}
	}
}
provider "template" {} 
`,
			Expected: helper.Issues{},
		},
		{
			Name: "required_providers string",
			Content: `
terraform {
	required_providers {
		template = "~> 2" 
	}
}
provider "template" {} 
`,
			Expected: helper.Issues{},
		},
		{
			Name: "required_providers object missing version",
			Content: `
terraform {
	required_providers {
		template = {
			source = "hashicorp/template"
		}
	}
}

provider "template" {} 
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `Missing version constraint for provider "template" in "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   10,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   10,
							Column: 20,
						},
					},
				},
			},
		},
		{
			Name: "single provider with alias",
			Content: `
provider "template" {
	alias = "b"
}
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `Missing version constraint for provider "template" in "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   2,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   2,
							Column: 20,
						},
					},
				},
			},
		},
		{
			Name: "version set",
			Content: `
terraform {
  required_providers {
    template = {
			source = "hashicorp/template"
			version = "~> 2"
		}
  }
}

provider "template" {
	version = "~> 2"
} 
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `provider version constraint should be specified via "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   11,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   11,
							Column: 20,
						},
					},
				},
			},
		},
		{
			Name: "version set with configuration_aliases",
			Content: `
terraform {
  required_providers {
    template = {
			source = "hashicorp/template"
			version = "~> 2"
			configuration_aliases = [template.alias]
		}
  }
}

data "template_file" "foo" {
	provider = template.alias
}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "version set with alias",
			Content: `
terraform {
  required_providers {
    template = {
			source = "hashicorp/template"
			version = "~> 2"
		}
  }
}

provider "template" {
	alias   = "foo"
	version = "~> 2"
} 
`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `provider version constraint should be specified via "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   11,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   11,
							Column: 20,
						},
					},
				},
			},
		},
		{
			Name: "terraform provider",
			Content: `
data "terraform_remote_state" "foo" {}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "builtin provider",
			Content: `
terraform {
	required_providers {
		test = {
			source = "terraform.io/builtin/test"
		}
	}
}
resource "test_assertions" "foo" {}
`,
			Expected: helper.Issues{},
		},
		{
			Name: "resource provider ref",
			Content: `
terraform {
  required_providers {
    google = {
      version = "~> 4.27.0"
	}
  }
}

resource "google_compute_instance" "foo" {
  provider = google-beta
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `Missing version constraint for provider "google-beta" in "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   10,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   10,
							Column: 41,
						},
					},
				},
			},
		},
		{
			Name: "resource provider ref as string",
			Content: `
terraform {
  required_providers {
    google = {
      version = "~> 4.27.0"
    }
  }
}

resource "google_compute_instance" "foo" {
  provider = "google-beta"
}`,
			Expected: helper.Issues{
				{
					Rule:    NewTerraformRequiredProvidersRule(),
					Message: `Missing version constraint for provider "google-beta" in "required_providers"`,
					Range: hcl.Range{
						Filename: "module.tf",
						Start: hcl.Pos{
							Line:   10,
							Column: 1,
						},
						End: hcl.Pos{
							Line:   10,
							Column: 41,
						},
					},
				},
			},
		},
	}

	rule := NewTerraformRequiredProvidersRule()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			runner := testRunner(t, map[string]string{"module.tf": tc.Content})

			if err := rule.Check(runner); err != nil {
				t.Fatalf("Unexpected error occurred: %s", err)
			}

			helper.AssertIssues(t, tc.Expected, runner.Runner.(*helper.Runner).Issues)
		})
	}
}
