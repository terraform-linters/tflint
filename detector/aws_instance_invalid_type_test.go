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

func TestDetectAwsInstanceInvalidType(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "t2.2xlarge is invalid",
			Src: `
resource "aws_instance" "web" {
    instance_type = "t2.2xlarge"
}`,
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "WARNING",
					Message: "\"t2.2xlarge\" is invalid instance type.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "t2.micro is valid",
			Src: `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
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
		d := &Detector{
			ListMap:    listMap,
			EvalConfig: evalConfig,
		}

		var issues = []*issue.Issue{}
		d.DetectAwsInstanceInvalidType(&issues)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", issues, tc.Issues, tc.Name)
		}
	}
}
