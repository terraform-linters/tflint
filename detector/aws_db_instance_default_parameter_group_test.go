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

func TestDetectAwsDBInstanceDefaultParameterGroup(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "default.mysql5.6 is default parameter group",
			Src: `
resource "aws_db_instance" "db" {
    parameter_group_name = "default.mysql5.6"
}`,
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "NOTICE",
					Message: "\"default.mysql5.6\" is default parameter group. You cannot edit it.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "application5.6 is not default parameter group",
			Src: `
resource "aws_db_instance" "db" {
    parameter_group_name = "application5.6"
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
		d := &AwsDBInstanceDefaultParameterGroupDetector{
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
