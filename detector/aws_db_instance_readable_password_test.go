package detector

import (
	"reflect"
	"testing"

	"github.com/k0kubun/pp"
	"github.com/wata727/tflint/config"
	"github.com/wata727/tflint/issue"
)

func TestDetectAwsDBInstanceReadablePassword(t *testing.T) {
	cases := []struct {
		Name   string
		Src    string
		Issues []*issue.Issue
	}{
		{
			Name: "write password directly",
			Src: `
resource "aws_db_instance" "mysql" {
    password = "super_secret"
}`,
			Issues: []*issue.Issue{
				&issue.Issue{
					Type:    "WARNING",
					Message: "Password for the master DB user is readable. recommend using environment variables.",
					Line:    3,
					File:    "test.tf",
				},
			},
		},
		{
			Name: "using environment variable",
			Src: `
resource "aws_db_instance" "mysql" {
    password = "${var.mysql_password}"
}`,
			Issues: []*issue.Issue{},
		},
	}

	for _, tc := range cases {
		var issues = []*issue.Issue{}
		TestDetectByCreatorName(
			"CreateAwsDBInstanceReadablePasswordDetector",
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
