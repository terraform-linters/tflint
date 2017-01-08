package detector

import (
	"reflect"
	"testing"

	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
)

func TestDetectAwsDBInstancePreviousType(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "db.t1.micro is previous type",
			Src: `
resource "aws_db_instance" "mysql" {
    instance_class = "db.t1.micro"
}`,
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "WARNING",
					Message: "\"db.t1.micro\" is previous generation instance type.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "db.t2.micro is not previous type",
			Src: `
resource "aws_db_instance" "mysql" {
    instance_class = "db.t2.micro"
}`,
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsDBInstancePreviousTypeDetector",
			tc.Src,
			config.Init(),
			config.Init().NewAwsClient(),
			&issues,
		)

		if !reflect.DeepEqual(issues, tc.Issues) {
			t.Fatalf("Bad: %s\nExpected: %s\n\ntestcase: %s", issues, tc.Issues, tc.Name)
		}
	}
}
