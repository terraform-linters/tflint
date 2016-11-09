package detector

import (
	"reflect"
	"testing"

	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/parser"
	"github.com/wata727/tflint/config"
	eval "github.com/wata727/tflint/evaluator"
	"github.com/wata727/tflint/issue"
)

func TestDetectAwsInstanceNotSpecifiedIamProfile(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "iam_instance_profile is not specified",
			Src: `
resource "aws_instance" "web" {
    instance_type = "t2.2xlarge"
}`,
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "NOTICE",
					Message: "\"iam_instance_profile\" is not specified. You cannot edit this value later.",
					Line:    2,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "iam_instance_profile is specified",
			Src: `
resource "aws_instance" "web" {
    instance_type = "t2.micro"
    iam_instance_profile = "test"
}`,
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		listMap := make(map[string]*ast.ObjectList)
		root, _ := parser.Parse([]byte(tc.Src))
		list, _ := root.Node.(*ast.ObjectList)
		listMap["test.tf"] = list

		evalConfig, _ := eval.NewEvaluator(listMap, config.Init("", ""))
		d := &Detector{
			ListMap:    listMap,
			EvalConfig: evalConfig,
		}

		var issues = []*issue.Issue{}
		d.DetectAwsInstanceNotSpecifiedIamProfile(&issues)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", issues, tc.Issues, tc.Name)
		}
	}
}
