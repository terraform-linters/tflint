package detector

import (
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
)

func TestDetectAwsElastiCacheClusterPreviousType(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "cache.t1.micro is previous type",
			Src: `
resource "aws_elasticache_cluster" "redis" {
    node_type = "cache.t1.micro"
}`,
			Issues: []*issue.Issue{
				{
					Detector: "aws_elasticache_cluster_previous_type",
					Type:     "WARNING",
					Message:  "\"cache.t1.micro\" is previous generation node type.",
					Line:     3,
					File:     "test.tf",
					Link:     "https://github.com/wata727/tflint/blob/master/docs/aws_elasticache_cluster_previous_type.md",
				},
			},
		},
		{
			Name: "cache.t2.micro is not previous type",
			Src: `
resource "aws_elasticache_cluster" "redis" {
    node_type = "cache.t2.micro"
}`,
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsElastiCacheClusterPreviousTypeDetector",
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
