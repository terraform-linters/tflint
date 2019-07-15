package client

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	awsbase "github.com/hashicorp/aws-sdk-go-base"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/hashicorp/terraform/configs"
	homedir "github.com/mitchellh/go-homedir"
)

func Test_getBaseConfig(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	home, err := homedir.Expand("~/")
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Name     string
		Creds    AwsCredentials
		File     string
		Expected *awsbase.Config
	}{
		{
			Name: "static credentials",
			Creds: AwsCredentials{
				AccessKey: "AWS_ACCESS_KEY",
				SecretKey: "AWS_SECRET_KEY",
				Region:    "us-east-1",
			},
			Expected: &awsbase.Config{
				AccessKey: "AWS_ACCESS_KEY",
				SecretKey: "AWS_SECRET_KEY",
				Region:    "us-east-1",
			},
		},
		{
			Name: "shared credentials",
			Creds: AwsCredentials{
				Profile:   "default",
				CredsFile: "~/.aws/creds",
				Region:    "us-east-1",
			},
			Expected: &awsbase.Config{
				Profile:       "default",
				CredsFilename: filepath.Join(home, ".aws", "creds"),
				Region:        "us-east-1",
			},
		},
		{
			Name: "static credentials provider",
			File: "static-creds.tf",
			Expected: &awsbase.Config{
				AccessKey: "AWS_ACCESS_KEY",
				SecretKey: "AWS_SECRET_KEY",
				Region:    "us-east-1",
			},
		},
		{
			Name: "shared credentials provider",
			File: "shared-creds.tf",
			Expected: &awsbase.Config{
				Profile:       "default",
				CredsFilename: filepath.Join(home, ".aws", "creds"),
				Region:        "us-east-1",
			},
		},
		{
			Name:     "assume role provider",
			File:     "assume-role.tf",
			Expected: &awsbase.Config{},
		},
		{
			Name: "prefer tflint static credentials over provider",
			File: "static-creds.tf",
			Creds: AwsCredentials{
				AccessKey: "TFLINT_AWS_ACCESS_KEY",
				SecretKey: "TFLINT_AWS_SECRET_KEY",
				Region:    "us-east-2",
			},
			Expected: &awsbase.Config{
				AccessKey: "TFLINT_AWS_ACCESS_KEY",
				SecretKey: "TFLINT_AWS_SECRET_KEY",
				Region:    "us-east-2",
			},
		},
		{
			Name: "prefer tflint shared credentials over provider",
			File: "shared-creds.tf",
			Creds: AwsCredentials{
				Profile:   "terraform",
				CredsFile: "~/.aws/tf_credentials",
				Region:    "us-east-2",
			},
			Expected: &awsbase.Config{
				Profile:       "terraform",
				CredsFilename: filepath.Join(home, ".aws", "tf_credentials"),
				Region:        "us-east-2",
			},
		},
	}

	for _, tc := range cases {
		var pc *configs.Provider
		if tc.File != "" {
			parser := hclparse.NewParser()
			f, diags := parser.ParseHCLFile(filepath.Join(currentDir, "test-fixtures", tc.File))
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			body, _, diags := f.Body.PartialContent(&hcl.BodySchema{
				Blocks: []hcl.BlockHeaderSchema{
					{
						Type:       "provider",
						LabelNames: []string{"name"},
					},
				},
			})
			if diags.HasErrors() {
				t.Fatal(diags)
			}

			pc = &configs.Provider{
				Config: body.Blocks[0].Body,
			}
		}

		base, err := getBaseConfig(pc, tc.Creds)
		if err != nil {
			t.Fatalf("Failed `%s` test: Unexpected error occurred: %s", tc.Name, err)
		}
		if !cmp.Equal(tc.Expected, base) {
			t.Fatalf("Failed `%s` test: Diff=%s", tc.Name, cmp.Diff(tc.Expected, base))
		}
	}
}
