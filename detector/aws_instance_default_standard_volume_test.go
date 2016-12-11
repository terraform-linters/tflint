package detector

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/evaluator"
	"github.com/wata727/tflint/issue"
)

func TestDetectAwsInstanceDefaultStandardVolume(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "volume_type is not specified in root_block_device",
			Src: `
resource "aws_instance" "web" {
    instance_type = "c3.2xlarge"

    root_block_device = {
        volume_size = "24"
    }
}`,
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "WARNING",
					Message: "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:    5,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "volume_type is not specified in ebs_block_device",
			Src: `
resource "aws_instance" "web" {
    instance_type = "c3.2xlarge"

    ebs_block_device = {
        volume_size = "24"
    }
}`,
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "WARNING",
					Message: "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:    5,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "volume_type is not specified in multi devices",
			Src: `
resource "aws_instance" "web" {
    instance_type = "c3.2xlarge"

    root_block_device = {
        volume_size = "100"
    }

    ebs_block_device = {
        volume_size = "24"
    }

    ebs_block_device = {
        volume_size = "10"
    }
}`,
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "WARNING",
					Message: "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:    5,
					File:    "test.tf",
				},
				&issue.Issue{
					Type:    "WARNING",
					Message: "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:    9,
					File:    "test.tf",
				},
				&issue.Issue{
					Type:    "WARNING",
					Message: "\"volume_type\" is not specified. Default standard volume type is not recommended. You can use \"gp2\", \"io1\", etc instead.",
					Line:    13,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "volume_type is specified",
			Src: `
resource "aws_instance" "web" {
    instance_type = "c3.2xlarge"

    root_block_device = {
        volume_type = "gp2"
        volume_size = "24"
    }
}`,
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		listMap := make(map[string]*ast.ObjectList)
		root, _ := parser.Parse([]byte(tc.Src))
		list, _ := root.Node.(*ast.ObjectList)
		listMap["test.tf"] = list

		evalConfig, _ := evaluator.NewEvaluator(listMap, config.Init())
		d := &AwsInstanceDefaultStandardVolumeDetector{
			&Detector{
				ListMap:    listMap,
				EvalConfig: evalConfig,
			},
		}

		var issues = []*issue.Issue{}
		d.Detect(&issues)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", issues, tc.Issues, tc.Name)
		}
	}
}
