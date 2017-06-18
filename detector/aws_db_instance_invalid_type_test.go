package detector

import (
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
)

func TestDetectAwsDBInstanceInvalidType(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "m4.2xlarge is invalid",
			Src: `
resource "aws_db_instance" "mysql" {
    instance_class = "m4.2xlarge"
}`,
			Issues: []*issue.Issue{
				{
					Type:    "ERROR",
					Message: "\"m4.2xlarge\" is invalid instance type.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "db.m4.2xlarge is valid",
			Src: `
resource "aws_db_instance" "mysql" {
    instance_class = "db.m4.2xlarge"
}`,
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		var issues = []*issue.Issue{}
		err := TestDetectByCreatorName(
			"CreateAwsDBInstanceInvalidTypeDetector",
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
