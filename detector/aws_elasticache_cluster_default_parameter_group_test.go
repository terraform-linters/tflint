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

func TestDetectAwsElastiCacheClusterDefaultParameterGroup(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "default.redis3.2 is default parameter group",
			Src: `
resource "aws_elasticache_cluster" "cache" {
    parameter_group_name = "default.redis3.2"
}`,
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "NOTICE",
					Message: "\"default.redis3.2\" is default parameter group. You cannot edit it.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "application3.2 is not default parameter group",
			Src: `
resource "aws_elasticache_cluster" "cache" {
    parameter_group_name = "application3.2"
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
		d := &AwsElastiCacheClusterDefaultParameterGroupDetector{
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
