package awsrules

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform/configs"
	"github.com/hashicorp/terraform/configs/configload"
	"github.com/hashicorp/terraform/terraform"
	"github.com/wata727/tflint/issue"
	"github.com/wata727/tflint/tflint"
)

func Test_AwsInstanceDefaultStandardVolume(t *testing.T) {
	cases := []struct {
		Name     string
		Content  string
		Expected issue.Issues
	}{
		{
			Name: "volume_type is not specified in root_block_device",
			Content: `
resource "aws_instance" "web" {
    instance_type = "c3.2xlarge"

    root_block_device = {
        volume_size = "24"
    }
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_instance_default_standard_volume",
					Type:     issue.WARNING,
					Message:  "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:     5,
					File:     "resource.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_default_standard_volume.md",
				},
			},
		},
		{
			Name: "volume_type is not specified in ebs_block_device",
			Content: `
resource "aws_instance" "web" {
    instance_type = "c3.2xlarge"

    ebs_block_device = {
        volume_size = "24"
    }
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_instance_default_standard_volume",
					Type:     issue.WARNING,
					Message:  "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:     5,
					File:     "resource.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_default_standard_volume.md",
				},
			},
		},
		{
			Name: "volume_type is specified",
			Content: `
resource "aws_instance" "web" {
    instance_type = "c3.2xlarge"

    root_block_device = {
        volume_type = "gp2"
        volume_size = "24"
    }
}`,
			Expected: []*issue.Issue{},
		},
		// TODO: An error occurred in Terraform v0.12. Attribute redefined; The argument "ebs_block_device" was already set

		// 		{
		// 			Name: "volume_type is not specified in multi devices",
		// 			Content: `
		// resource "aws_instance" "web" {
		//     instance_type = "c3.2xlarge"
		//     ami = "ami-1234567"

		//     root_block_device = {
		//         volume_size = "100"
		//     }

		//     ebs_block_device = {
		//         device_name = "foo"
		//         volume_size = "24"
		//     }

		//     ebs_block_device = {
		//         device_name = "bar"
		//         volume_size = "10"
		//     }
		// }`,
		// 			Expected: []*issue.Issue{
		// 				{
		// 					Detector: "aws_instance_default_standard_volume",
		// 					Type:     issue.WARNING,
		// 					Message:  "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
		// 					Line:     6,
		// 					File:     "resource.tf",
		// 					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_default_standard_volume.md",
		// 				},
		// 				{
		// 					Detector: "aws_instance_default_standard_volume",
		// 					Type:     issue.WARNING,
		// 					Message:  "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
		// 					Line:     10,
		// 					File:     "resource.tf",
		// 					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_default_standard_volume.md",
		// 				},
		// 				{
		// 					Detector: "aws_instance_default_standard_volume",
		// 					Type:     issue.WARNING,
		// 					Message:  "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
		// 					Line:     15,
		// 					File:     "resource.tf",
		// 					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_instance_default_standard_volume.md",
		// 				},
		// 			},
		// 		},
	}

	dir, err := ioutil.TempDir("", "AwsInstanceDefaultStandardVolume")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	for _, tc := range cases {
		loader, err := configload.NewLoader(&configload.Config{})
		if err != nil {
			t.Fatal(err)
		}

		err = ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		mod, diags := loader.Parser().LoadConfigDir(dir)
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			t.Fatal(tfdiags)
		}

		runner := tflint.NewRunner(tflint.EmptyConfig(), cfg, map[string]*terraform.InputValue{})
		rule := NewAwsInstanceDefaultStandardVolumeRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
