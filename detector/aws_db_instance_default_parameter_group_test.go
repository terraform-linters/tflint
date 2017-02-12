package detector

import (
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
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
		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsDBInstanceDefaultParameterGroupDetector",
			tc.Src,
			"",
			config.Init(),
			config.Init().NewAwsClient(),
			&issues,
		)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("\nBad: %s\nExpected: %s\n\ntestcase: %s", pp.Sprint(issues), pp.Sprint(tc.Issues), tc.Name)
		}
	}
}
