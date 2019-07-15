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
	"github.com/wata727/tflint/project"
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
					Link:     project.ReferenceLink("aws_instance_default_standard_volume"),
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
					Link:     project.ReferenceLink("aws_instance_default_standard_volume"),
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
		{
			Name: "volume_type is not specified in multi devices",
			Content: `
resource "aws_instance" "web" {
    instance_type = "c3.2xlarge"
    ami = "ami-1234567"

    root_block_device {
        volume_size = "100"
    }

    ebs_block_device {
        device_name = "foo"
        volume_size = "24"
    }

    ebs_block_device {
        device_name = "bar"
        volume_size = "10"
    }
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_instance_default_standard_volume",
					Type:     issue.WARNING,
					Message:  "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:     6,
					File:     "resource.tf",
					Link:     project.ReferenceLink("aws_instance_default_standard_volume"),
				},
				{
					Detector: "aws_instance_default_standard_volume",
					Type:     issue.WARNING,
					Message:  "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:     10,
					File:     "resource.tf",
					Link:     project.ReferenceLink("aws_instance_default_standard_volume"),
				},
				{
					Detector: "aws_instance_default_standard_volume",
					Type:     issue.WARNING,
					Message:  "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:     15,
					File:     "resource.tf",
					Link:     project.ReferenceLink("aws_instance_default_standard_volume"),
				},
			},
		},
		{
			Name: "volume_type is null",
			Content: `
variable "volume_type" {
	type    = string
	default = null
}

resource "aws_instance" "web" {
    instance_type = "c3.2xlarge"

    ebs_block_device {
        volume_type = var.volume_type
        volume_size = "24"
    }
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_instance_default_standard_volume",
					Type:     issue.WARNING,
					Message:  "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:     11,
					File:     "resource.tf",
					Link:     project.ReferenceLink("aws_instance_default_standard_volume"),
				},
			},
		},
		{
			Name: "volume_type attribute is null",
			Content: `
variable "volume_type" {
	type    = string
	default = null
}

resource "aws_instance" "web" {
    instance_type = "c3.2xlarge"

    ebs_block_device = {
        volume_type = var.volume_type
        volume_size = "24"
    }
}`,
			Expected: []*issue.Issue{},
		},
		{
			Name: "dynamic blocks",
			Content: `
variable "volumes" {
	type    = list(string)
	default = ["100", "200"]
}

resource "aws_instance" "web" {
    instance_type = "c3.2xlarge"

    dynamic "ebs_block_device" {
		for_each = var.volumes

		content {
			volume_size = ebs_block_device.value
		}
	}
}`,
			Expected: []*issue.Issue{
				{
					Detector: "aws_instance_default_standard_volume",
					Type:     issue.WARNING,
					Message:  "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:     13,
					File:     "resource.tf",
					Link:     project.ReferenceLink("aws_instance_default_standard_volume"),
				},
			},
		},
	}

	dir, err := ioutil.TempDir("", "AwsInstanceDefaultStandardVolume")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(currentDir)

	err = os.Chdir(dir)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		loader, err := configload.NewLoader(&configload.Config{})
		if err != nil {
			t.Fatal(err)
		}

		err = ioutil.WriteFile(dir+"/resource.tf", []byte(tc.Content), os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		mod, diags := loader.Parser().LoadConfigDir(".")
		if diags.HasErrors() {
			t.Fatal(diags)
		}
		cfg, tfdiags := configs.BuildConfig(mod, configs.DisabledModuleWalker)
		if tfdiags.HasErrors() {
			t.Fatal(tfdiags)
		}

		runner, err := tflint.NewRunner(tflint.EmptyConfig(), map[string]tflint.Annotations{}, cfg, map[string]*terraform.InputValue{})
		if err != nil {
			t.Fatal(err)
		}
		rule := NewAwsInstanceDefaultStandardVolumeRule()

		if err = rule.Check(runner); err != nil {
			t.Fatalf("Unexpected error occurred: %s", err)
		}

		if !cmp.Equal(tc.Expected, runner.Issues) {
			t.Fatalf("Expected issues are not matched:\n %s\n", cmp.Diff(tc.Expected, runner.Issues))
		}
	}
}
